package helper

import (
	"reflect"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/paybyphone/phpipam-sdk-go/controllers/subnets"
)

func TestSubnetsSorter(t *testing.T) {
	in := SubnetsSorter{
		subnets.Subnet{
			SubnetAddress: "10.10.3.0",
			Mask:          24,
		},
		subnets.Subnet{
			SubnetAddress: "10.10.1.0",
			Mask:          24,
		},
		subnets.Subnet{
			SubnetAddress: "10.0.0.0",
			Mask:          8,
		},
		subnets.Subnet{
			SubnetAddress: "172.16.1.0",
			Mask:          24,
		},
		subnets.Subnet{
			SubnetAddress: "10.10.4.0",
			Mask:          24,
		},
		subnets.Subnet{
			SubnetAddress: "172.16.0.0",
			Mask:          12,
		},
	}
	expected := SubnetsSorter{
		subnets.Subnet{
			SubnetAddress: "10.0.0.0",
			Mask:          8,
		},
		subnets.Subnet{
			SubnetAddress: "10.10.1.0",
			Mask:          24,
		},
		subnets.Subnet{
			SubnetAddress: "10.10.3.0",
			Mask:          24,
		},
		subnets.Subnet{
			SubnetAddress: "10.10.4.0",
			Mask:          24,
		},
		subnets.Subnet{
			SubnetAddress: "172.16.0.0",
			Mask:          12,
		},
		subnets.Subnet{
			SubnetAddress: "172.16.1.0",
			Mask:          24,
		},
	}

	actual := in
	sort.Sort(actual)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected %s, got %s", spew.Sdump(expected), spew.Sdump(actual))
	}
}
