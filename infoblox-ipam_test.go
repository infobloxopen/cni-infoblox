package ibcni

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/containernetworking/cni/pkg/types"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

type MockObjectManager struct {
	nameArg                           string
	globalNetviewArg, localNetviewArg string
	getNetviewArg, createNetviewArg   string

	netviewArg, cidrArg   string
	eaArg                 ibclient.EA
	ipAddrArg, macAddrArg string
	prefixLenArg          uint
	vmIDArg               string
	networkRefArg         string
	eadefArg              ibclient.EADefinition

	getNetworkView, createNetworkView     *ibclient.NetworkView
	globalNetviewRef, localNetviewRef     string
	network                               *ibclient.Network
	getFixedAddress, allocateFixedAddress *ibclient.FixedAddress
	eaDefinition                          *ibclient.EADefinition
	fixedAddressRef, networkRef           string
	err                                   error

	createNetworkViewCalled, createNetworkCalled, allocateIPCalled bool
	getNetworkNilEaReturnsNil                                      bool
	getNetworkReturnsNil                                           bool

	networkContainerPoolArgs                                              []string
	allocateNetworkReturns                                                []*ibclient.Network
	createNetworkContainerReturns                                         []*ibclient.NetworkContainer
	getNetworkContainerReturns                                            []*ibclient.NetworkContainer
	getNetworkContainerCnt, createNetworkContainerCnt, allocateNetworkCnt int
}

func (f *MockObjectManager) CreateNetworkView(name string) (*ibclient.NetworkView, error) {
	Expect(name).To(Equal(f.createNetviewArg))

	f.createNetworkViewCalled = true

	return f.createNetworkView, f.err
}

func (f *MockObjectManager) CreateDefaultNetviews(globalNetview string, localNetview string) (globalNetviewRef string, localNetviewRef string, err error) {
	Expect(globalNetview).To(Equal(f.globalNetviewArg))
	Expect(localNetview).To(Equal(f.localNetviewArg))

	return f.globalNetviewRef, f.localNetviewRef, f.err
}

func (f *MockObjectManager) CreateNetwork(netview string, cidr string, name string) (*ibclient.Network, error) {
	Expect(netview).To(Equal(f.netviewArg))
	Expect(cidr).To(Equal(f.cidrArg))
	Expect(name).To(Equal(f.nameArg))

	f.createNetworkCalled = true

	return f.network, f.err
}

func (f *MockObjectManager) CreateNetworkContainer(netview string, cidr string) (*ibclient.NetworkContainer, error) {
	Expect(netview).To(Equal(f.netviewArg))
	Expect(cidr).To(Equal(f.networkContainerPoolArgs[f.createNetworkContainerCnt]))

	networkContainer := f.createNetworkContainerReturns[f.createNetworkContainerCnt]
	f.createNetworkContainerCnt++

	return networkContainer, f.err
}

func (f *MockObjectManager) GetNetworkView(name string) (*ibclient.NetworkView, error) {
	Expect(name).To(Equal(f.getNetviewArg))

	return f.getNetworkView, f.err
}

func (f *MockObjectManager) GetNetwork(netview string, cidr string, ea ibclient.EA) (*ibclient.Network, error) {
	Expect(netview).To(Equal(f.netviewArg))

	if cidr == "" {
		Expect(ea).To(Equal(f.eaArg))
	}

	if ea == nil {
		Expect(cidr).To(Equal(f.cidrArg))
	}

	if f.getNetworkReturnsNil || (f.getNetworkNilEaReturnsNil && ea == nil) {
		return nil, f.err
	}
	return f.network, f.err
}

func (f *MockObjectManager) GetNetworkContainer(netview string, cidr string) (*ibclient.NetworkContainer, error) {
	Expect(netview).To(Equal(f.netviewArg))
	Expect(cidr).To(Equal(f.networkContainerPoolArgs[f.getNetworkContainerCnt]))

	networkContainer := f.getNetworkContainerReturns[f.getNetworkContainerCnt]
	f.getNetworkContainerCnt++

	return networkContainer, f.err
}

func (f *MockObjectManager) AllocateIP(netview string, cidr string, ipAddr string, macAddr string, vmID string) (*ibclient.FixedAddress, error) {
	Expect(netview).To(Equal(f.netviewArg))
	Expect(cidr).To(Equal(f.cidrArg))
	Expect(ipAddr).To(Equal(f.ipAddrArg))
	Expect(macAddr).To(Equal(f.macAddrArg))
	Expect(vmID).To(Equal(f.vmIDArg))

	f.allocateIPCalled = true

	return f.allocateFixedAddress, f.err
}

func (f *MockObjectManager) AllocateNetwork(netview string, cidr string, prefixLen uint, name string) (*ibclient.Network, error) {
	Expect(netview).To(Equal(f.netviewArg))
	Expect(cidr).To(Equal(f.networkContainerPoolArgs[f.allocateNetworkCnt]))
	Expect(prefixLen).To(Equal(f.prefixLenArg))
	Expect(name).To(Equal(f.nameArg))

	network := f.allocateNetworkReturns[f.allocateNetworkCnt]
	f.allocateNetworkCnt++

	return network, f.err
}

func (f *MockObjectManager) GetFixedAddress(netview string, cidr string, ipAddr string, macAddr string) (*ibclient.FixedAddress, error) {
	Expect(netview).To(Equal(f.netviewArg))
	Expect(cidr).To(Equal(f.cidrArg))
	Expect(ipAddr).To(Equal(f.ipAddrArg))
	Expect(macAddr).To(Equal(f.macAddrArg))

	return f.getFixedAddress, f.err
}

func (f *MockObjectManager) ReleaseIP(netview string, cidr string, ipAddr string, macAddr string) (string, error) {
	Expect(netview).To(Equal(f.netviewArg))
	Expect(cidr).To(Equal(f.cidrArg))
	Expect(ipAddr).To(Equal(f.ipAddrArg))
	Expect(macAddr).To(Equal(f.macAddrArg))

	return f.fixedAddressRef, f.err
}

func (f *MockObjectManager) DeleteNetwork(networkRef string, netview string) (string, error) {
	Expect(networkRef).To(Equal(f.networkRefArg))
	Expect(netview).To(Equal(f.netviewArg))

	return f.networkRef, f.err
}

func (f *MockObjectManager) GetEADefinition(name string) (*ibclient.EADefinition, error) {
	Expect(name).To(Equal(f.nameArg))

	return f.eaDefinition, f.err
}

func (f *MockObjectManager) CreateEADefinition(eadef ibclient.EADefinition) (*ibclient.EADefinition, error) {
	Expect(eadef).To(Equal(f.eadefArg))

	return f.eaDefinition, f.err
}

var _ = Describe("InfobloxIpam", func() {
	log.SetOutput(ioutil.Discard)

	defaultNetworkView := "default-view"
	defaultNetworkContainer := "192.168.100.0/24"
	defaultPrefixLen := uint(24)

	Describe("RequestNetworkView", func() {
		Context("When requested Network View already exists", func() {
			testView := "test-view"
			testNetView := &ibclient.NetworkView{Name: testView}

			objMgr := &MockObjectManager{
				getNetviewArg:  testView,
				getNetworkView: testNetView,
				err:            nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var netview string
			var err error
			It("Should pass expected netviewName to ObjectManager.GetNetworkView", func() {
				netview, err = ibDriver.RequestNetworkView(testView)
			})
			It("Should not call ObjectManager.CreateNetworkView", func() {
				Expect(objMgr.createNetworkViewCalled).To(BeFalse())
			})
			It("Should return expected NetworkView Name", func() {
				Expect(netview).To(Equal(testView))
				Expect(err).To(BeNil())
			})
		})

		Context("When requested Network View does not already exist", func() {
			testView := "test-view"
			testNetView := &ibclient.NetworkView{Name: testView}

			objMgr := &MockObjectManager{
				getNetviewArg:     testView,
				getNetworkView:    nil,
				createNetviewArg:  testView,
				createNetworkView: testNetView,
				err:               nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var netview string
			var err error
			It("Should pass expected netviewName to ObjectManager.GetNetworkView and ObjectManager.CreateNetworkView", func() {
				netview, err = ibDriver.RequestNetworkView(testView)
			})
			It("Should call ObjectManager.CreateNetworkView", func() {
				Expect(objMgr.createNetworkViewCalled).To(BeTrue())
			})
			It("Should return created NetworkView Name", func() {
				Expect(netview).To(Equal(testView))
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("RequestAddress", func() {
		Context("When requested Fixed Address already exists", func() {
			testView := "test-view"
			testCidr := "192.168.10.0/24"
			testIpAddr := "192.168.10.10"
			testMacAddr := "11:22:33:44:55:66"
			testVmID := "1234567890abcdef"

			testFixedAddr := &ibclient.FixedAddress{
				NetviewName: testView,
				Cidr:        testCidr,
				IPAddress:   testIpAddr,
				Mac:         testMacAddr,
				Ea:          ibclient.EA{"VM ID": testVmID},
			}

			objMgr := &MockObjectManager{
				netviewArg: testView,
				cidrArg:    testCidr,
				ipAddrArg:  testIpAddr,
				macAddrArg: testMacAddr,
				vmIDArg:    testVmID,

				getFixedAddress: testFixedAddr,
				err:             nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var ipAddr string
			var err error
			It("Should pass expected arguments to ObjectManager.RequestAddress", func() {
				ipAddr, err = ibDriver.RequestAddress(testView, testCidr, testIpAddr, testMacAddr, testVmID)
			})
			It("Should not call ObjectManager.AllocateIP", func() {
				Expect(objMgr.allocateIPCalled).To(BeFalse())
			})
			It("Should return expected FixedAddress object", func() {
				Expect(ipAddr).To(Equal(testIpAddr))
				Expect(err).To(BeNil())
			})
		})

		Context("When requested Fixed Address does not already exist", func() {
			testView := "test-view"
			testCidr := "192.168.10.0/24"
			testIpAddr := "192.168.10.10"
			testMacAddr := "11:22:33:44:55:66"
			testVmID := "1234567890abcdef"

			testFixedAddr := &ibclient.FixedAddress{
				NetviewName: testView,
				Cidr:        testCidr,
				IPAddress:   testIpAddr,
				Mac:         testMacAddr,
				Ea:          ibclient.EA{"VM ID": testVmID},
			}

			objMgr := &MockObjectManager{
				netviewArg: testView,
				cidrArg:    testCidr,
				ipAddrArg:  testIpAddr,
				macAddrArg: testMacAddr,
				vmIDArg:    testVmID,

				getFixedAddress:      nil,
				allocateFixedAddress: testFixedAddr,
				err:                  nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var ipAddr string
			var err error
			It("Should pass expected arguments to ObjectManager.RequestAddress and ObjectManager.AllocateIP", func() {
				ipAddr, err = ibDriver.RequestAddress(testView, testCidr, testIpAddr, testMacAddr, testVmID)
			})
			It("Should call ObjectManager.AllocateIP", func() {
				Expect(objMgr.allocateIPCalled).To(BeTrue())
			})
			It("Should return expected FixedAddress object", func() {
				Expect(ipAddr).To(Equal(testIpAddr))
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("ReleaseAddress", func() {
		testView := ""
		testIpAddr := "192.168.10.10"
		testMacAddr := "11:22:33:44:55:66"
		testIpRef := "fixedaddress/ZG5zLmJpbmRfY25h:192.168.10.10/default-view"

		objMgr := &MockObjectManager{
			netviewArg: defaultNetworkView,
			ipAddrArg:  testIpAddr,
			macAddrArg: testMacAddr,

			fixedAddressRef: testIpRef,
			err:             nil,
		}

		ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

		var ipRef string
		var err error
		It("Should pass expected arguments to ObjectManager.ReleaseIP", func() {
			ipRef, err = ibDriver.ReleaseAddress(testView, testIpAddr, testMacAddr)
		})
		It("Should return expected FixedAddress ref", func() {
			Expect(ipRef).To(Equal(testIpRef))
			Expect(err).To(BeNil())
		})
	})

	Describe("requestSpecificNetwork", func() {
		Context("When network with matching cidr and name already exist", func() {
			testView := "test-view"
			testCidr := "192.168.10.0/24"
			testNetworkName := "yellow"
			testEa := ibclient.EA{"Network Name": testNetworkName}

			testNetwork := &ibclient.Network{
				NetviewName: testView,
				Cidr:        testCidr,
				Ea:          testEa,
			}

			objMgr := &MockObjectManager{
				netviewArg: testView,
				cidrArg:    testCidr,

				network: testNetwork,
				err:     nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var network *ibclient.Network
			var err error
			It("Should pass expected arguments to ObjectManager.GetNetwork", func() {
				network, err = ibDriver.requestSpecificNetwork(testView, testCidr, testNetworkName)
			})
			It("Should not call ObjectManager.CreateNetwork", func() {
				Expect(objMgr.createNetworkCalled).To(BeFalse())
			})
			It("Should return expected Network object", func() {
				Expect(network).To(Equal(testNetwork))
				Expect(err).To(BeNil())
			})
		})

		Context("When no matching network exists", func() {
			testView := "test-view"
			testCidr := "192.168.10.0/24"
			testNetworkName := "yellow"
			testEa := ibclient.EA{"Network Name": testNetworkName}

			testNetwork := &ibclient.Network{
				NetviewName: testView,
				Cidr:        testCidr,
				Ea:          testEa,
			}

			objMgr := &MockObjectManager{
				netviewArg: testView,
				cidrArg:    testCidr,
				eaArg:      testEa,
				nameArg:    testNetworkName,

				network: testNetwork,
				err:     nil,

				getNetworkReturnsNil: true,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var network *ibclient.Network
			var err error
			It("Should pass expected arguments to ObjectManager.GetNetwork", func() {
				network, err = ibDriver.requestSpecificNetwork(testView, testCidr, testNetworkName)
			})
			It("Should not call ObjectManager.CreateNetwork", func() {
				Expect(objMgr.createNetworkCalled).To(BeTrue())
			})
			It("Should return nil Network object", func() {
				Expect(network).To(Equal(testNetwork))
				Expect(err).To(BeNil())
			})
		})

		Context("When network with same name exists but has different cidr", func() {
			testView := "test-view"
			testCidr := "192.168.10.0/24"
			mismatchCidr := "192.168.20.0/24"
			testNetworkName := "yellow"
			testEa := ibclient.EA{"Network Name": testNetworkName}

			testNetwork := &ibclient.Network{
				NetviewName: testView,
				Cidr:        mismatchCidr,
				Ea:          testEa,
			}

			objMgr := &MockObjectManager{
				netviewArg: testView,
				cidrArg:    testCidr,
				eaArg:      testEa,

				network: testNetwork,
				err:     nil,

				getNetworkNilEaReturnsNil: true,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var network *ibclient.Network
			var err error
			It("Should pass expected arguments to ObjectManager.GetNetwork", func() {
				network, err = ibDriver.requestSpecificNetwork(testView, testCidr, testNetworkName)
			})
			It("Should not call ObjectManager.CreateNetwork", func() {
				Expect(objMgr.createNetworkCalled).To(BeFalse())
			})
			It("Should return nil Network object", func() {
				Expect(network).To(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("allocateNetworkHelper", func() {
		Context("When a network can be allocated from a network container", func() {
			testView := "test-view"
			testContainerArr := []string{"192.168.10.0/24", "192.168.20.0/24"}
			testPrefixLen := uint(26)
			testNetworkName := "yellow"

			testContainers := strings.Join(testContainerArr, ",")

			testNetwork := &ibclient.Network{
				NetviewName: testView,
				Cidr:        "192.168.20.100/26",
			}

			objMgr := &MockObjectManager{
				netviewArg:   testView,
				prefixLenArg: testPrefixLen,
				nameArg:      testNetworkName,

				networkContainerPoolArgs: testContainerArr,
				allocateNetworkReturns: []*ibclient.Network{
					nil,
					testNetwork,
				},
				createNetworkContainerReturns: []*ibclient.NetworkContainer{
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[0],
					},
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[1],
					},
				},
				getNetworkContainerReturns: []*ibclient.NetworkContainer{
					nil,
					nil,
				},

				err: nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, testView, testContainers, testPrefixLen)

			var network *ibclient.Network
			var err error
			It("Should pass expected arguments to ObjectManager.GetNetworkContainer/CreateNetworkContainer/AllocateNetwork", func() {
				network, err = ibDriver.allocateNetworkHelper(testView, testPrefixLen, testNetworkName)
			})
			It("Should call Object Manager the expected no. of times", func() {
				Expect(objMgr.getNetworkContainerCnt).To(Equal(2))
				Expect(objMgr.createNetworkContainerCnt).To(Equal(2))
				Expect(objMgr.allocateNetworkCnt).To(Equal(2))
			})
			It("Should return expected Network object", func() {
				Expect(network).To(Equal(testNetwork))
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("allocateNetwork", func() {
		Context("When no network can be allocated from all Network Containers", func() {
			testView := "test-view"
			testContainerArr := []string{"192.168.10.0/24", "192.168.20.0/24"}
			testPrefixLen := uint(26)
			testNetworkName := "yellow"

			testContainers := strings.Join(testContainerArr, ",")

			objMgr := &MockObjectManager{
				netviewArg:   testView,
				prefixLenArg: testPrefixLen,
				nameArg:      testNetworkName,

				networkContainerPoolArgs: append(testContainerArr, testContainerArr...),
				allocateNetworkReturns: []*ibclient.Network{
					nil,
					nil,
					nil,
					nil,
				},
				createNetworkContainerReturns: []*ibclient.NetworkContainer{
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[0],
					},
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[1],
					},
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[0],
					},
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[1],
					},
				},
				getNetworkContainerReturns: []*ibclient.NetworkContainer{
					nil,
					nil,
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[0],
					},
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[1],
					},
				},

				err: nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, testView, testContainers, testPrefixLen)

			var network *ibclient.Network
			var err error
			It("Should pass expected arguments to ObjectManager.GetNetworkContainer/CreateNetworkContainer/AllocateNetwork", func() {
				network, err = ibDriver.allocateNetwork(testPrefixLen, testNetworkName)
			})
			It("Should call Object Manager the expected no. of times", func() {
				Expect(objMgr.getNetworkContainerCnt).To(Equal(2))
				Expect(objMgr.createNetworkContainerCnt).To(Equal(2))
				Expect(objMgr.allocateNetworkCnt).To(Equal(4))
			})
			It("Should return expected Network object", func() {
				Expect(network).To(BeNil())
				Expect(err.Error()).To(Equal("Cannot allocate network in Address Space"))
			})
		})
	})

	Describe("RequestNetwork", func() {
		Context("When called with a specific subnet cidr", func() {
			testView := "test-view"
			testIp := net.IPv4(byte(192), byte(168), byte(30), byte(0))
			testMask := net.IPv4Mask(byte(255), byte(255), byte(255), byte(0))
			testIPNet := net.IPNet{IP: testIp, Mask: testMask}
			testCidr := testIPNet.String()
			testNetworkName := "yellow"
			testEa := ibclient.EA{"Network Name": testNetworkName}

			testNetwork := &ibclient.Network{
				NetviewName: testView,
				Cidr:        testCidr,
				Ea:          testEa,
			}

			netconf := NetConfig{
				Name: testNetworkName,
				IPAM: &IPAMConfig{
					NetworkView: testView,
					Subnet: types.IPNet{
						IP:   testIp,
						Mask: testMask,
					},
				},
			}

			objMgr := &MockObjectManager{
				netviewArg:   testView,
				prefixLenArg: defaultPrefixLen,
				nameArg:      testNetworkName,

				networkContainerPoolArgs: []string{defaultNetworkContainer},
				allocateNetworkReturns: []*ibclient.Network{
					testNetwork,
				},
				createNetworkContainerReturns: []*ibclient.NetworkContainer{
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        defaultNetworkContainer,
					},
				},
				getNetworkContainerReturns: []*ibclient.NetworkContainer{
					nil,
				},

				cidrArg: testCidr,
				eaArg:   testEa,
				network: testNetwork,
				err:     nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, defaultNetworkView, defaultNetworkContainer, defaultPrefixLen)

			var network string
			var err error
			It("Should pass expected arguments to ObjectManager.GetNetworkContainer/CreateNetworkContainer/AllocateNetwork", func() {
				network, err = ibDriver.RequestNetwork(netconf)
			})
			It("Should call Object Manager the expected no. of times", func() {
				Expect(objMgr.getNetworkContainerCnt).To(Equal(0))
				Expect(objMgr.createNetworkContainerCnt).To(Equal(0))
				Expect(objMgr.allocateNetworkCnt).To(Equal(0))
			})
			It("Should return expected Network object", func() {
				Expect(network).To(Equal(testCidr))
				Expect(err).To(BeNil())
			})
		})

		Context("When a network can be allocated from a network container", func() {
			testView := "test-view"
			testContainerArr := []string{"192.168.10.0/24"}
			testPrefixLen := uint(26)
			testNetworkName := "yellow"
			testCidr := "192.168.10.100/26"
			testEa := ibclient.EA{"Network Name": testNetworkName}

			testContainers := strings.Join(testContainerArr, ",")

			netconf := NetConfig{
				Name: testNetworkName,
				IPAM: &IPAMConfig{
					NetworkView: testView,
					Subnet:      types.IPNet{},
				},
			}

			testNetwork := &ibclient.Network{
				NetviewName: testView,
				Cidr:        testCidr,
			}

			objMgr := &MockObjectManager{
				netviewArg:   testView,
				prefixLenArg: testPrefixLen,
				nameArg:      testNetworkName,

				networkContainerPoolArgs: testContainerArr,
				allocateNetworkReturns: []*ibclient.Network{
					testNetwork,
				},
				createNetworkContainerReturns: []*ibclient.NetworkContainer{
					&ibclient.NetworkContainer{
						NetviewName: testView,
						Cidr:        testContainerArr[0],
					},
				},
				getNetworkContainerReturns: []*ibclient.NetworkContainer{
					nil,
				},

				eaArg: testEa,
				err:   nil,
			}

			ibDriver := NewInfobloxDriver(objMgr, testView, testContainers, testPrefixLen)

			var network string
			var err error
			It("Should pass expected arguments to ObjectManager.GetNetworkContainer/CreateNetworkContainer/AllocateNetwork", func() {
				network, err = ibDriver.RequestNetwork(netconf)
			})
			It("Should call Object Manager the expected no. of times", func() {
				Expect(objMgr.getNetworkContainerCnt).To(Equal(1))
				Expect(objMgr.createNetworkContainerCnt).To(Equal(1))
				Expect(objMgr.allocateNetworkCnt).To(Equal(1))
			})
			It("Should return expected Network object", func() {
				Expect(network).To(Equal(testCidr))
				Expect(err).To(BeNil())
			})
		})
	})
})
