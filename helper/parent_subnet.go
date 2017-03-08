package helper

import (
	"fmt"
	"net"

	"github.com/paybyphone/phpipam-sdk-go/controllers/subnets"
	"github.com/paybyphone/phpipam-sdk-go/phpipam/session"
	"github.com/sirupsen/logrus"
)

// ParentSubnetIDForCIDR finds the parent subnet ID for a specific address and
// mask in PHPIPAM via API.
//
// We decrement our subnet mask until we get to 8, which is the largets block
// allocation allowed by the IANA (aka a Class A).
//
// 0 is returned if no subnet is found.
func ParentSubnetIDForCIDR(session *session.Session, addr string, mask int) int {
	logrus.Debugf("Looking for parent subnet for CIDR %s/%d", addr, mask)

	c := subnets.NewController(session)

	// Start the counter at the netmask one more bit "wider" than the original
	// mask, so that we don't find the child subnet instead.
	n := mask - 1
	for n >= 8 {
		_, net, err := net.ParseCIDR(fmt.Sprintf("%s/%d", addr, n))
		if err != nil {
			logrus.Fatalf("Error parsing subnet/CIDR %s/%d: %s", addr, mask, err)
		}
		logrus.Debugf("Looking for subnet CIDR %s in new PHPIPAM database", net.String())
		subnets, err := c.GetSubnetsByCIDR(net.String())
		switch {
		case err == nil:
			logrus.Debugf("Parent found: subnet ID %d for CIDR %s in new PHPIPAM database", subnets[0].ID, net.String())
			return subnets[0].ID
		case err.Error() == "Error from API (404): No subnets found":
			logrus.Debugf("Subnet %s not found in PHPIPAM", net.String())
			n--
			continue
		default:
			logrus.Fatalf("Error searching for subnet: %s", err)
		}
	}
	return 0
}
