package helper

import (
	"fmt"

	"github.com/paybyphone/phpipam-sdk-go/controllers/subnets"
)

// SubnetsSorter implements a sorter for the subnets.Subnet data type. This is
// so that we can sort the order that subnets get added, so that we can
// determine master subnet properly.
type SubnetsSorter []subnets.Subnet

// Len implements sort.Interface.Len for SubnetsSorter.
func (s SubnetsSorter) Len() int {
	return len(s)
}

// Swap implements sort.Interface.Swap for SubnetsSorter.
func (s SubnetsSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements sort.Interface.Less for SubnetsSorter. The IP/mask is
// converted to string form and then sorted lexicograpically.
func (s SubnetsSorter) Less(i, j int) bool {
	return fmt.Sprintf("%s/%d", s[i].SubnetAddress, s[i].Mask) < fmt.Sprintf("%s/%d", s[j].SubnetAddress, s[j].Mask)
}
