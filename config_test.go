package ibcni

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os"
	"strconv"
	"strings"
)

var _ = Describe("LoadConfig", func() {

	It("Should return expected config according to command line", func() {
		const (
			GridHost         = "192.168.124.200"
			WapiPort         = "443"
			WapiUsername     = "ibadmin"
			WapiPassword     = "ibpassword"
			WapiVersion      = "2.0"
			SocketDir        = "/run/cni"
			DriverName       = "infoblox"
			SslVerify        = "false"
			NetworkView      = "default"
			NetworkContainer = "192.168.0.0/24,192.169.0.0/24"
			PrefixLength     = "25"
		)

		cmdLine := fmt.Sprintf("infoblox-cni-daemon --grid-host=%s --wapi-port=%s --wapi-username=%s --wapi-password=%s --wapi-version=%s --socket-dir=%s --driver-name=%s --ssl-verify=%s --network-view=%s --network-container=%s --prefix-length=%s",
			GridHost, WapiPort, WapiUsername, WapiPassword, WapiVersion,
			SocketDir, DriverName, SslVerify, NetworkView, NetworkContainer, PrefixLength)

		os.Args = strings.Split(cmdLine, " ")

		config := LoadConfig()

		Expect(config.GridHost).To(Equal(GridHost))
		Expect(config.WapiPort).To(Equal(WapiPort))
		Expect(config.WapiUsername).To(Equal(WapiUsername))
		Expect(config.WapiPassword).To(Equal(WapiPassword))
		Expect(config.WapiVer).To(Equal(WapiVersion))
		Expect(config.SocketDir).To(Equal(SocketDir))
		Expect(config.DriverName).To(Equal(DriverName))
		Expect(config.SslVerify).To(Equal(SslVerify))
		Expect(config.NetworkView).To(Equal(NetworkView))
		Expect(config.NetworkContainer).To(Equal(NetworkContainer))
		prefixLen, _ := strconv.ParseUint(PrefixLength, 10, 64)
		Expect(config.PrefixLength).To(Equal(uint(prefixLen)))
	})
})
