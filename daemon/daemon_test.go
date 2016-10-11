package main

import (
	"github.com/containernetworking/cni/pkg/types"
	. "github.com/infobloxopen/cni-infoblox"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

type MockInfobloxDriver struct {
	netviewNameArg, cidrArg, ipAddrArg, macAddrArg, vmIDArg string

	netconfArg NetConfig

	requestNetworkViewRet, requestAddressRet, releaseAddressRet, requestNetworkRet string

	requestNetworkViewCnt, requestAddressCnt, releaseAddressCnt, requestNetworkCnt int

	err error
}

func (ibDrv *MockInfobloxDriver) RequestNetworkView(netviewName string) (string, error) {
	Expect(netviewName).To(Equal(ibDrv.netviewNameArg))

	ibDrv.requestNetworkViewCnt++

	return ibDrv.requestNetworkViewRet, ibDrv.err
}

func (ibDrv *MockInfobloxDriver) RequestAddress(netviewName string, cidr string, ipAddr string, macAddr string, vmID string) (string, error) {
	Expect(netviewName).To(Equal(ibDrv.netviewNameArg))
	Expect(cidr).To(Equal(ibDrv.cidrArg))
	Expect(ipAddr).To(Equal(ibDrv.ipAddrArg))
	Expect(macAddr).To(Equal(ibDrv.macAddrArg))
	Expect(vmID).To(Equal(ibDrv.vmIDArg))

	ibDrv.requestAddressCnt++

	return ibDrv.requestAddressRet, ibDrv.err
}

func (ibDrv *MockInfobloxDriver) ReleaseAddress(netviewName string, ipAddr string, macAddr string) (string, error) {
	Expect(netviewName).To(Equal(ibDrv.netviewNameArg))
	Expect(ipAddr).To(Equal(ibDrv.ipAddrArg))
	Expect(macAddr).To(Equal(ibDrv.macAddrArg))

	ibDrv.releaseAddressCnt++

	return ibDrv.releaseAddressRet, ibDrv.err
}

func (ibDrv *MockInfobloxDriver) RequestNetwork(netconf NetConfig) (string, error) {
	Expect(netconf).To(Equal(ibDrv.netconfArg))

	ibDrv.requestNetworkCnt++

	return ibDrv.requestNetworkRet, ibDrv.err
}

var _ = Describe("Daemon", func() {
	log.SetOutput(ioutil.Discard)

	testNetworkName := "yellow"
	testIpamType := "infoblox"
	testView := "test-view"

	testIp := net.IPv4(byte(192), byte(168), byte(30), byte(0))
	testMask := net.IPv4Mask(byte(255), byte(255), byte(255), byte(0))
	testIPNet := net.IPNet{IP: testIp, Mask: testMask}
	testCidr := testIPNet.String()

	testContainerID := "abcdef123456"
	testIfMac := "11:22:33:44:55:66"

	testAllocatedIPStr := "192.168.30.21"
	testAllocatedIP := net.ParseIP(testAllocatedIPStr)
	testAllocatedIPNet := net.IPNet{IP: testAllocatedIP, Mask: testMask}

	testIpamConf := fmt.Sprintf(`
{
    "name": "%s",
    "ipam": {
        "type": "%s",
		"network-view": "%s",
        "subnet": "%s"
    }
}`, testNetworkName, testIpamType, testView, testCidr)

	netconf := NetConfig{
		Name: testNetworkName,
		IPAM: &IPAMConfig{
			Type:        testIpamType,
			NetworkView: testView,
			Subnet: types.IPNet{
				IP:   testIp,
				Mask: testMask,
			},
		},
	}

	Context("Allocate Method", func() {
		ibDriver := &MockInfobloxDriver{
			netviewNameArg: testView,
			netconfArg:     netconf,
			cidrArg:        testCidr,
			ipAddrArg:      "",
			macAddrArg:     testIfMac,
			vmIDArg:        testContainerID,

			requestNetworkViewRet: testView,
			requestNetworkRet:     testCidr,
			requestAddressRet:     testAllocatedIPStr,
		}

		ib := newInfoblox(ibDriver)

		args := &ExtCmdArgs{}
		args.ContainerID = testContainerID
		args.IfMac = testIfMac
		args.StdinData = []byte(testIpamConf)

		allocateResult := &types.Result{}

		var err error
		It("Should pass expected arguments to InfobloxDriver methods", func() {
			err = ib.Allocate(args, allocateResult)
		})
		It("Should call InfobloxDriver methods the expected no. of times", func() {
			Expect(ibDriver.requestNetworkViewCnt).To(Equal(1))
			Expect(ibDriver.requestNetworkCnt).To(Equal(1))
			Expect(ibDriver.requestAddressCnt).To(Equal(1))
		})
		It("Should return the expected result", func() {
			Expect(err).To(BeNil())
			Expect(allocateResult.IP4.IP).To(Equal(testAllocatedIPNet))
		})
	})

	Context("Release Method", func() {
		testAddrRef := "fixedaddress/ZG5zLmJpbmRfY25h:192.168.30.21/test-view"

		ibDriver := &MockInfobloxDriver{
			netviewNameArg: testView,
			ipAddrArg:      "",
			macAddrArg:     testIfMac,

			releaseAddressRet: testAddrRef,
		}

		ib := newInfoblox(ibDriver)

		args := &ExtCmdArgs{}
		args.ContainerID = testContainerID
		args.IfMac = testIfMac
		args.StdinData = []byte(testIpamConf)

		var err error
		It("Should pass expected arguments to InfobloxDriver methods", func() {
			err = ib.Release(args, nil)
		})
		It("Should call InfobloxDriver methods the expected no. of times", func() {
			Expect(ibDriver.releaseAddressCnt).To(Equal(1))
		})
		It("Should return the expected result", func() {
			Expect(err).To(BeNil())
		})
	})

	Context("getInfobloxDriver", func() {
		containersArr := []string{"192.168.0.0/24", "192.169.0.0/24"}

		config := &Config{}
		config.GridHost = "192.168.124.200"
		config.WapiPort = "443"
		config.WapiUsername = "ibadmin"
		config.WapiPassword = "ibpassword"
		config.WapiVer = "2.0"
		config.SocketDir = "/run/cni"
		config.DriverName = "infoblox"
		config.SslVerify = "false"
		config.NetworkView = "default"
		config.NetworkContainer = strings.Join(containersArr, ",")
		config.PrefixLength = uint(26)

		ibDrv := getInfobloxDriver(config)

		It("Should initialize driver with expected values", func() {
			Expect(ibDrv.DefaultNetworkView).To(Equal(config.NetworkView))
			Expect(ibDrv.DefaultPrefixLen).To(Equal(config.PrefixLength))
			Expect(len(ibDrv.Containers)).To(Equal(len(containersArr)))
			for i, c := range ibDrv.Containers {
				Expect(c.NetworkContainer).To(Equal(containersArr[i]))
			}
		})
	})

})
