// Package subnets provides types and methods for working with the subnets
// controller.
package subnets

import (
	"fmt"

	"github.com/paybyphone/phpipam-sdk-go/phpipam"
	"github.com/paybyphone/phpipam-sdk-go/phpipam/client"
	"github.com/paybyphone/phpipam-sdk-go/phpipam/session"
)

// Subnet represents a PHPIPAM subnet.
type Subnet struct {
	// The subnet ID.
	ID int `json:"id,string,omitempty"`

	// The subnet address, in dotted quad format (i.e. A.B.C.D).
	SubnetAddress string `json:"subnet,omitempty"`

	// The subnet's mask in number of bits (i.e. 24).
	Mask int `json:"mask,string,omitempty"`

	// A detailed description of the subnet.
	Description string `json:"description,omitempty"`

	// The section ID to add the subnet to (required when adding).
	SectionID int `json:"sectionId,string,omitempty"`

	// The ID of a linked IPv6 subnet.
	LinkedSubnet int `json:"linked_subnet,string,omitempty"`

	// The ID of the VLAN that this subnet belongs to.
	VLANID int `json:"vlanId,string,omitempty"`

	// The ID of the VRF this subnet belongs to.
	VRFID int `json:"vrfId,string,omitempty"`

	// The parent subnet ID if this is a nested subnet.
	MasterSubnetID int `json:"masterSubnetId,string,omitempty"`

	// The ID of the nameserver to attache the subnet to.
	NameserverID int `json:"nameserverId,string,omitempty"`

	// true if the name should be displayed in listing instead of the subnet
	// address.
	ShowName phpipam.BoolIntString `json:"showName,omitempty"`

	// A JSON object, stringified, that represents the permissions for this
	// section.
	Permissions string `json:"permissions,omitempty"`

	// Controls if PTR records should be created for the subnet.
	DNSRecursive phpipam.BoolIntString `json:"DNSrecursive,omitempty"`

	// Controls if DNS hostname records are displayed.
	DNSRecords phpipam.BoolIntString `json:"DNSrecords,omitempty"`

	// Controls if IP requests are allowed for the subnet.
	AllowRequests phpipam.BoolIntString `json:"allowRequests,omitempty"`

	// The ID of the scan agent to use for the subnet.
	ScanAgent int `json:"scanAgent,string,omitempty"`

	// Controls if the subnet should be included in status checks.
	PingSubnet phpipam.BoolIntString `json:"pingSubnet,omitempty"`

	// Controls if new hosts should be discovered for new host scans.
	DiscoverSubnet phpipam.BoolIntString `json:"discoverSubnet,omitempty"`

	// Controls if we are adding a subnet or folder.
	IsFolder phpipam.BoolIntString `json:"isFolder,omitempty"`

	// Marks the subnet as used.
	IsFull phpipam.BoolIntString `json:"isFull,omitempty"`

	// The tag ID for the subnet.
	State int `json:"state,string,omitempty"`

	// The threshold of the subnet.
	Threshold int `json:"threshold,string,omitempty"`

	// The location index of the subnet.
	Location int `json:"location,string,omitempty"`

	// The date of the last edit to this resource.
	EditDate string `json:"editDate,omitempty"`
}

// Controller is the base client for the Subnets controller.
type Controller struct {
	client.Client
}

// NewController returns a new instance of the client for the Subnets controller.
func NewController(sess *session.Session) *Controller {
	c := &Controller{
		Client: *client.NewClient(sess),
	}
	return c
}

// CreateSubnet creates a subnet by sending a POST request.
func (c *Controller) CreateSubnet(in Subnet) (message string, err error) {
	err = c.SendRequest("POST", "/subnets/", &in, &message)
	return
}

// GetSubnetByID GETs a subnet via its ID.
func (c *Controller) GetSubnetByID(id int) (out Subnet, err error) {
	err = c.SendRequest("GET", fmt.Sprintf("/subnets/%d/", id), &struct{}{}, &out)
	return
}

// GetSubnetsByCIDR GETs a subnet via its CIDR (i.e. 10.10.1.0/24).
//
// The function's name reflects the fact that an array of subnets is returned
// through the API, although it remains unclear how to actually query this
// method in a way that would return multiple results. Using a broader CIDR
// will not return multiple results, and using the CIDR of a master subnet will
// return that subnet only.
func (c *Controller) GetSubnetsByCIDR(cidr string) (out []Subnet, err error) {
	err = c.SendRequest("GET", fmt.Sprintf("/subnets/cidr/%s/", cidr), &struct{}{}, &out)
	return
}

// UpdateSubnet updates a subnet by sending a PATCH request.
//
// Note you cannot use this function to update a subnet's CIDR - to split,
// grow, or renumber a subnet, you need to use other methods that are currently
// not implemented in this SDK. See the API spec for more details.
func (c *Controller) UpdateSubnet(in Subnet) (message string, err error) {
	err = c.SendRequest("PATCH", "/subnets/", &in, &message)
	return
}

// DeleteSubnet deletes a subnet by its ID.
func (c *Controller) DeleteSubnet(id int) (message string, err error) {
	err = c.SendRequest("DELETE", fmt.Sprintf("/subnets/%d/", id), &struct{}{}, &message)
	return
}
