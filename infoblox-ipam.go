package main

import (
	"errors"
	"fmt"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"log"
	"net"
	//	"strconv"
	"strings"
)

type Container struct {
	NetworkContainer string // CIDR of Network Container
	ContainerObj     *ibclient.NetworkContainer
	exhausted        bool
}

type InfobloxDriver struct {
	objMgr       *ibclient.ObjectManager
	networkView  string
	prefixLength uint
	containers   []Container
}

func (ibDrv *InfobloxDriver) RequestAddress(netviewName string, cidr string, macAddr string, vmID string) (string, error) {
	if len(macAddr) == 0 {
		log.Printf("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.\n")
	}
	fixedAddr, _ := ibDrv.objMgr.AllocateIP(netviewName, cidr, macAddr, vmID)

	fmt.Printf("RequestAddress: fixedAddr is '%s'\n", *fixedAddr)
	return fmt.Sprintf("%s", fixedAddr.IPAddress), nil
}

func (ibDrv *InfobloxDriver) ReleaseAddress(netviewName string, ipAddr string, macAddr string) (ref string, err error) {
	if netviewName == "" {
		netviewName = ibDrv.networkView
	}
	ref, err = ibDrv.objMgr.ReleaseIP(netviewName, ipAddr, macAddr)
	if ref == "" {
		log.Printf("***** IP Cannot be deleted '%s', '%s', '%s'! *******\n", netviewName, ipAddr, macAddr)
	}

	return
}

func (ibDrv *InfobloxDriver) createNetworkContainer(netview string, pool string) (*ibclient.NetworkContainer, error) {
	container, err := ibDrv.objMgr.GetNetworkContainer(netview, pool)
	if container == nil {
		container, err = ibDrv.objMgr.CreateNetworkContainer(netview, pool)
	}

	return container, err
}

func (ibDrv *InfobloxDriver) nextAvailableContainer() *Container {
	for i, _ := range ibDrv.containers {
		if !ibDrv.containers[i].exhausted {
			return &ibDrv.containers[i]
		}
	}

	return nil
}

func (ibDrv *InfobloxDriver) resetContainers() {
	for i, _ := range ibDrv.containers {
		ibDrv.containers[i].exhausted = false
	}
}

func (ibDrv *InfobloxDriver) allocateNetworkHelper(netview string, prefixLen uint, name string) (network *ibclient.Network, err error) {
	fmt.Printf("allocateNetworkHelper(): netview is '%s',  prefixLen is '%d', name is '%s'\n", netview, prefixLen, name)
	container := ibDrv.nextAvailableContainer()
	for container != nil {
		log.Printf("Allocating network from Container:'%s'", container.NetworkContainer)
		if container.ContainerObj == nil {
			var err error
			container.ContainerObj, err = ibDrv.createNetworkContainer(netview, container.NetworkContainer)
			if err != nil {
				return nil, err
			}
		}
		network, err = ibDrv.objMgr.AllocateNetwork(netview, container.NetworkContainer, prefixLen, name)
		if network != nil {
			break
		}
		container.exhausted = true
		container = ibDrv.nextAvailableContainer()
	}

	return network, nil
}

func (ibDrv *InfobloxDriver) allocateNetwork(prefixLen uint, name string) (network *ibclient.Network, err error) {
	fmt.Printf("allocateNetwork(): prefixLen is '%d', name is '%s'\n", prefixLen, name)
	if prefixLen == 0 {
		prefixLen = ibDrv.prefixLength
	}
	network, err = ibDrv.allocateNetworkHelper(ibDrv.networkView, prefixLen, name)
	if network == nil {
		ibDrv.resetContainers()
		network, err = ibDrv.allocateNetworkHelper(ibDrv.networkView, prefixLen, name)
	}

	if network == nil {
		err = errors.New("Cannot allocate network in Address Space")
	}
	return
}

func (ibDrv *InfobloxDriver) requestSpecificNetwork(netview string, subnet string, name string) (*ibclient.Network, error) {
	network, err := ibDrv.objMgr.GetNetwork(netview, subnet, nil)
	if err != nil {
		return nil, err
	}
	if network != nil {
		fmt.Printf("GetNetwork: network is '%s'\n", *network)
		if n, ok := network.Ea["Network Name"]; !ok || n != name {
			fmt.Printf("GetNetwork: network is already used '%s'\n", *network)
			return nil, nil
		}
	} else {
		networkByName, err := ibDrv.objMgr.GetNetwork(netview, "", ibclient.EA{"Network Name": name})
		if err != nil {
			return nil, err
		}
		if networkByName != nil {
			fmt.Printf("GetNetworkByName: network is '%s'\n", *networkByName)
			if networkByName.Cidr != subnet {
				fmt.Printf("GetNetworkByName: network name has different Cidr '%s'\n", networkByName.Cidr)
				return nil, nil
			}
		}
	}

	if network == nil {
		network, err = ibDrv.objMgr.CreateNetwork(netview, subnet, name)
		fmt.Printf("CreateNetwork: network is '%s', err is '%s'\n", *network, err)
	}

	return network, err
}

/*
func (ibDrv *InfobloxDriver) RequestNetwork(netviewName string, cidr string, prefixLength uint, gw string) (network string, gw string,  err error) {
	var ibNetwork *ibclient.Network
	if cidr != "" {
		ibNetwork, err = ibDrv.requestSpecificNetwork(netviewName, cidr)
	} else {
		if netviewName == ibDrv.networkView {
			ibNetwork, err = ibDrv.allocateNetwork(prefixLength)
		} else {
			log.Printf("Incorrect Network View name specified: '%s'", netviewName)
			return "", "", nil
		}
	}

	fmt.Printf("RequestNetwork: network is '%s'\n", ibNetwork)
	res = ""
	if ibNetwork != nil {
		res = ibNetwork.Cidr
	}

	return res, gw, err
}
*/

func (ibDrv *InfobloxDriver) RequestNetwork(netconf NetConfig) (network string, gw string, err error) {
	var ibNetwork *ibclient.Network
	netviewName := netconf.IPAM.NetworkView
	cidr := net.IPNet{}
	fmt.Printf("RequestNetwork(): cidr is '%s', IPAM.Subnet is '%s'\n", cidr, netconf.IPAM.Subnet)
	if netconf.IPAM.Subnet.IP != nil {
		fmt.Printf("Subnet.IP is NOT nil!!!!!!!\n")
		cidr = net.IPNet{IP: netconf.IPAM.Subnet.IP, Mask: netconf.IPAM.Subnet.Mask}
		ibNetwork, err = ibDrv.requestSpecificNetwork(netviewName, cidr.String(), netconf.Name)
	} else {
		networkByName, err := ibDrv.objMgr.GetNetwork(netviewName, "", ibclient.EA{"Network Name": netconf.Name})
		if err != nil {
			return "", "", err
		}
		if networkByName != nil {
			fmt.Printf("GetNetworkByName: network is '%s'\n", *networkByName)
			ibNetwork = networkByName
		} else {
			if netviewName == "" {
				netviewName = ibDrv.networkView
			}
			if netviewName == ibDrv.networkView {
				prefixLen := ibDrv.prefixLength
				if netconf.IPAM.PrefixLength != 0 {
					prefixLen = netconf.IPAM.PrefixLength
				}
				ibNetwork, err = ibDrv.allocateNetwork(prefixLen, netconf.Name)
			} else {
				log.Printf("Incorrect Network View name specified: '%s'")
				return "", "", nil
			}
		}
	}

	fmt.Printf("RequestNetwork: network is '%s'\n", ibNetwork)
	res := ""
	if ibNetwork != nil {
		res = ibNetwork.Cidr
	}

	return res, gw, err
}

func makeContainers(containerList string) []Container {
	containers := make([]Container, 0)

	parts := strings.Split(containerList, ",")
	for _, p := range parts {
		containers = append(containers, Container{p, nil, false})
	}

	return containers
}

func NewInfobloxDriver(objMgr *ibclient.ObjectManager, networkView string, networkContainer string, prefixLength uint) *InfobloxDriver {
	return &InfobloxDriver{
		objMgr:       objMgr,
		networkView:  networkView,
		prefixLength: prefixLength,
		containers:   makeContainers(networkContainer),
	}
}
