package ibcni_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCniInfoblox(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CniInfoblox Suite")
}
