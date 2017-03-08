// Package main provides the main program for the PHPIPAM migrator.
//
// This migrator takes a PHPIPAM MySQL database from a pre-1.0 version (namely
// 0.8), queries its VLANs, subnets, and IP addresses, and migrates them to a
// recent PHPIPAM (1.2) via the API. The object here is to support a migration
// path from a version of PHPIPAM that's too old to have a database upgrade
// path.
package main

import (
	"database/sql"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	_ "github.com/go-sql-driver/mysql"

	"github.com/paybyphone/phpipam-legacy-migrator/helper"
	"github.com/paybyphone/phpipam-sdk-go/controllers/addresses"
	"github.com/paybyphone/phpipam-sdk-go/controllers/subnets"
	"github.com/paybyphone/phpipam-sdk-go/controllers/vlans"
	"github.com/paybyphone/phpipam-sdk-go/phpipam"
	"github.com/paybyphone/phpipam-sdk-go/phpipam/session"
	"github.com/sirupsen/logrus"
)

var (
	// ipamSession provides a PHPIPAM session. This is a global to make sure that
	// the same session token can be re-used across queries.
	ipamSession *session.Session

	// dbHost is the hostname housing the legacy DB. This deafults to blank,
	// which will use localhost.
	dbHost string

	// dbUser is the username to use when connecting to the legacy DB.
	dbUser string

	// dbPassword is the password to use when connecting to the legacy DB.
	dbPassword string

	// dbName is the database name to use when connecting to the legacy DB.
	dbName string

	// iapmAppID is the application ID for the new PHPIPAM API endpoint the tool
	// contacts. This is set up in the console. It can also be specified via the
	// PHPIPAM_APP_ID environment variable, and defaults to "default".
	ipamAppID string

	// ipamEndpoint is the full address to the location of the new PHPIPAM
	// endpoint (ie: https://phpipam.example.com/api). This can be also specified
	// with the PHPIPAM_ENDPOINT_ADDR environment variable. It defaults to
	// http://localhost/api if not specified.
	ipamEndpoint string

	// ipamPassword is the password for the PHPIPAM user that will be used to
	// contact the new endpoint. It can also be specified by the PHPIPAM_PASSWORD
	// environment variable.
	ipamPassword string

	// ipamUser is the user name that will be used to contact the new PHPIPAM
	// endpoint. It can be also specified via the PHPIPAM_USER_NAME environment
	// variable, and defaults to Admin.
	ipamUser string

	// The section ID to add the found subnets to. On a current default PHPIPAM
	// installation, the "Customers" section seems to be 1, so this is the
	// default.
	sectionID int

	// debug enables debug logging.
	debug bool
)

func init() {
	flag.StringVar(&dbHost, "dbhost", "", "The database host to connect to")
	flag.StringVar(&dbUser, "dbuser", "phpipam", "The database user to use")
	flag.StringVar(&dbPassword, "dbpassword", "", "The password for the database user")
	flag.StringVar(&dbName, "dbname", "phpipam", "The name of the database to import data from")
	flag.StringVar(&ipamAppID, "appid", "", "The PHPIPAM application ID to use")
	flag.StringVar(&ipamEndpoint, "endpoint", "", "The PHPIPAM endpoint to connect to")
	flag.StringVar(&ipamPassword, "password", "", "The password for the PHPIPAM user")
	flag.StringVar(&ipamUser, "user", "", "The user to use when connecting to PHPIPAM")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.IntVar(&sectionID, "sectionid", 1, "The section ID to add addresses to")

	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if dbPassword == "" {
		if dbHost != "" {
			// we only use TCP, and port 3306, so update the hostname so that it
			// works with the DSN.
			dbHost = fmt.Sprintf("tcp(%s:3306)", dbHost)
		}
		fmt.Printf("Enter the database password for %s@%s/%s: ", dbUser, dbHost, dbName)
		b, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			logrus.Fatalf("Error reading database user password: %s", err)
		}
		dbPassword = string(b)
	}

	if ipamPassword == "" && os.Getenv("PHPIPAM_PASSWORD") == "" {
		fmt.Print("Enter the PHPIPAM password:")
		b, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			logrus.Fatalf("Error reading PHPIPAM password: %s", err)
		}
		ipamPassword = string(b)
	}

	// set up the PHPIPAM connection
	ipamSession = session.NewSession(
		phpipam.Config{
			AppID:    ipamAppID,
			Endpoint: ipamEndpoint,
			Password: ipamPassword,
			Username: ipamUser,
		},
	)
}

// runSQL is a helper function that runs SQL. It logs the query as a debug
// message, and exits the program if the SQL query fails.
func runSQL(conn *sql.DB, query string) *sql.Rows {
	logrus.Debugf("Running SQL query: %s", query)
	rows, err := conn.Query(query)
	if err != nil {
		logrus.Fatalf("Fatal: error running SQL query: %s", err)
	}
	return rows
}

// decimalIPAddrToString converts a decimal IPv4 address to a dotted-quad
// string, ie: 1.2.3.4.
func decimalIPAddrToString(addr string) (string, error) {
	d, err := strconv.ParseUint(addr, 10, 32)
	if err != nil {
		return "", err
	}
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(d))
	return ip.String(), nil
}

// vlanIDForNumber fetches the VLAN ID for a specific VLAN number.
func vlanIDForNumber(n int) int {
	c := vlans.NewController(ipamSession)
	vlans, err := c.GetVLANsByNumber(n)
	if err != nil {
		logrus.Fatalf("Error getting VLAN ID for number %d: %s", n, err)
	}

	logrus.Debugf("Found VLAN ID %d for VLAN number %d in new PHPIPAM database", vlans[0].ID, n)
	return vlans[0].ID
}

// subnetIDForCIDR fetches a subnet ID via its CIDR subnet address
func subnetIDForCIDR(cidr string) int {
	c := subnets.NewController(ipamSession)
	subnets, err := c.GetSubnetsByCIDR(cidr)
	if err != nil {
		logrus.Fatalf("Error getting subnet ID for CIDR %s: %s", cidr, err)
	}

	logrus.Debugf("Found subnet ID %d for CIDR %s in new PHPIPAM database", subnets[0].ID, cidr)
	return subnets[0].ID
}

// fetchVLANs gets all the VLANs from the legacy DB and returns a []vlans.VLAN.
func fetchVLANs(conn *sql.DB) (out []vlans.VLAN) {
	logrus.Info("Fetching VLANs from legacy DB")

	rows := runSQL(conn, "select name, number, description from vlans")
	defer rows.Close()
	for rows.Next() {
		var name, description string
		var number int
		if err := rows.Scan(&name, &number, &description); err != nil {
			logrus.Fatalf("Error reading VLAN rows: %s", err)
		}
		out = append(out, vlans.VLAN{
			Name:        name,
			Number:      number,
			Description: description,
		})
		logrus.Debugf("Found VLAN - Name: %s, Number: %d, Description: %s", name, number, description)
	}
	if err := rows.Err(); err != nil {
		logrus.Fatalf("Error reading VLAN rows: %s", err)
	}
	logrus.Infof("Found %d VLANs to migrate", len(out))
	return
}

// fetchSubnets gets all of the IPv4 subnets from the legacy DB and returns a
// []subnets.Subnet.
//
// The SQL query joins 2 tables - subnets and vlans, to ensure that VLAN ID
// entries in the table are translated to their names, so that we can add the
// subnets to the VLANs in the new PHPIPAM instance by name.
func fetchSubnets(conn *sql.DB) (out []subnets.Subnet) {
	logrus.Info("Fetching subnets from legacy DB")

	rows := runSQL(conn, "select subnets.subnet, subnets.mask, subnets.description, vlans.number from subnets left join vlans on subnets.vlanId = vlans.vlanId")
	defer rows.Close()
	for rows.Next() {
		var mask int
		var vlanNumber sql.NullInt64
		var addr, description string
		if err := rows.Scan(&addr, &mask, &description, &vlanNumber); err != nil {
			logrus.Fatalf("Error reading subnet rows: %s", err)
		}

		// Our IP address is in decimal format, and needs converting to IPv4. If
		// this is an IPv6 address, we ignore the row.
		strAddr, err := decimalIPAddrToString(addr)
		if err != nil {
			logrus.Debugf("Ignoring inconvertible decimal address %s - possibly not an IPv4 address (%s)", addr, err)
			continue
		}

		var vlanID int
		if vlanNumber.Int64 != 0 {
			vlanID = vlanIDForNumber(int(vlanNumber.Int64))
		}

		out = append(out, subnets.Subnet{
			SubnetAddress: strAddr,
			Mask:          mask,
			Description:   description,
			VLANID:        vlanID,
			SectionID:     sectionID,
		})
		logrus.Debugf("Found subnet - Name: %s, Mask: %d, Description: %s, VLAN: %d", strAddr, mask, description, vlanNumber.Int64)
	}
	if err := rows.Err(); err != nil {
		logrus.Fatalf("Error reading subnet rows: %s", err)
	}

	logrus.Infof("Found %d subnets to migrate", len(out))
	return
}

// fetchAddresses gets all of the IPv4 addresses from the legacy DB and returns
// an []addresses.Address.
//
// The SQL query joins 2 tables - addresses and subnets, to ensure that we know
// what subnet that the IP address belongs to, without knowing its specific ID
// in the database.
func fetchAddresses(conn *sql.DB) (out []addresses.Address) {
	logrus.Info("Fetching addresses from legacy DB")

	rows := runSQL(conn, "select ipaddresses.ip_addr, ipaddresses.description, ipaddresses.dns_name, ipaddresses.note, subnets.subnet, subnets.mask from ipaddresses left join subnets on ipaddresses.subnetId=subnets.id")
	defer rows.Close()
	for rows.Next() {
		var ipAddr, description, dnsName, note, subnetAddr string
		var subnetMask int

		if err := rows.Scan(&ipAddr, &description, &dnsName, &note, &subnetAddr, &subnetMask); err != nil {
			logrus.Fatalf("Error reading address rows: %s", err)
		}

		// We have addresses that need converting to string format. Do this now.
		ipString, err := decimalIPAddrToString(ipAddr)
		if err != nil {
			logrus.Debugf("Ignoring inconvertible decimal IP address %s - possibly not an IPv4 address (%s)", ipAddr, err)
			continue
		}
		subnetString, err := decimalIPAddrToString(subnetAddr)
		if err != nil {
			logrus.Debugf("Ignoring inconvertible decimal subnet address %s - possibly not an IPv4 address (%s)", subnetAddr, err)
			continue
		}

		out = append(out, addresses.Address{
			SubnetID:    subnetIDForCIDR(fmt.Sprintf("%s/%d", subnetString, subnetMask)),
			IPAddress:   ipString,
			Description: description,
			Hostname:    dnsName,
			Note:        note,
		})
		logrus.Debugf("Found IP address - Address: %s, Description: %s, Hostname: %s, Note: %s, Subnet: %s/%d", ipString, description, dnsName, note, subnetString, subnetMask)
	}
	if err := rows.Err(); err != nil {
		logrus.Fatalf("Error reading address rows: %s", err)
	}
	logrus.Infof("Found %d addresses to migrate", len(out))
	return
}

// addVLANs adds the VLANs found into the new PHPIPAM instance.
func addVLANs(lans []vlans.VLAN) {
	logrus.Info("Adding VLANs.")

	c := vlans.NewController(ipamSession)
	for _, v := range lans {
		if _, err := c.CreateVLAN(v); err != nil {
			logrus.Fatalf("Error adding VLAN number %d: %s", v.Number, err)
		}
		logrus.Infof("VLAN number %d added successfully", v.Number)
	}
}

// addSubnets adds the subnets found into the new PHPIPAM instance.
//
// As the subnets are being added, we also check to see if we can find a parent
// subnet. In order to do this, the subnets are sorted first by way of
// SubnetsSorter, after which the list is iterated on.
func addSubnets(nets []subnets.Subnet) {
	data := helper.SubnetsSorter(nets)
	sort.Sort(data)

	logrus.Info("Adding subnets.")

	c := subnets.NewController(ipamSession)

	for _, v := range data {
		if id := helper.ParentSubnetIDForCIDR(ipamSession, v.SubnetAddress, v.Mask); id != 0 {
			v.MasterSubnetID = id
		}
		if _, err := c.CreateSubnet(v); err != nil {
			logrus.Fatalf("Error creating subnet %s/%d: %s", v.SubnetAddress, v.Mask, err)
		}
		logrus.Infof("Subnet address %s/%d added successfully", v.SubnetAddress, v.Mask)
	}
}

// addAddresses adds the IP addresses found into the new PHPIPAM instance.
func addAddresses(addrs []addresses.Address) {
	logrus.Info("Adding IP addresses.")

	c := addresses.NewController(ipamSession)
	for _, v := range addrs {
		if _, err := c.CreateAddress(v); err != nil {
			logrus.Fatalf("Error adding IP address %s: %s", v.IPAddress, err)
		}
		logrus.Infof("IP address %s added successfully", v.IPAddress)
	}
}

// connectDB sets up the database connection.
func connectDB() *sql.DB {
	logrus.Debugf("Connecting to DB: %s:[hidden]@%s/%s", dbUser, dbHost, dbName)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", dbUser, dbPassword, dbHost, dbName))
	if err != nil {
		logrus.Fatalf("Error configuring DB handle for %s:[hidden]@%s/%s: %s", dbUser, dbHost, dbName, err)
	}
	if err := db.Ping(); err != nil {
		logrus.Fatalf("Error connecting to DB %s:[hidden]@%s/%s: %s", dbUser, dbHost, dbName, err)
	}
	return db
}

func main() {
	logrus.Info("Migration starting.")

	db := connectDB()
	addVLANs(fetchVLANs(db))
	addSubnets(fetchSubnets(db))
	addAddresses(fetchAddresses(db))

	logrus.Info("Migration completed.")
}
