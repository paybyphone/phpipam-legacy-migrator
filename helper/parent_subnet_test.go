package helper

import (
	"os"
	"testing"

	"github.com/paybyphone/phpipam-sdk-go/phpipam/session"
	"github.com/paybyphone/phpipam-sdk-go/testacc"
	"github.com/sirupsen/logrus"
)

// TestParentSubnetIDForCIDR tests to see if we can find the parent subnet ID for 10.10.2.0/24 on a default PHPIPAM installation. This parent subnet is 10.10.0.0/16, which is subnet ID 2.
func TestParentSubnetIDForCIDR(t *testing.T) {
	testacc.VetAccConditions(t)
	sess := session.NewSession()

	expected := 2
	actual := ParentSubnetIDForCIDR(sess, "10.10.2.0", 24)

	if expected != actual {
		t.Fatalf("Expected master subnet ID to be %d, got %d", expected, actual)
	}
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	os.Exit(m.Run())
}
