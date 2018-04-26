// Copyright 2016 Infoblox Inc.
// All Rights Reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package ibcni

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	ibclient "github.com/infobloxopen/infoblox-go-client"
)

type Container struct {
	NetworkContainer string // CIDR of Network Container
	NetworkView      string // Network view
	ContainerObj     *ibclient.NetworkContainer
	exhausted        bool
}

type IBInfobloxDriver interface {
	RequestNetworkView(netviewName string) (string, error)
	RequestAddress(netviewName string, cidr string, ipAddr string, macAddr string, name string, vmID string) (string, error)
	GetAddress(netviewName string, cidr string, ipAddr string, macAddr string) (*ibclient.FixedAddress, error)
	UpdateAddress(fixedAddrRef string, macAddr string, name string, vmID string) (*ibclient.FixedAddress, error)
	ReleaseAddress(netviewName string, ipAddr string, macAddr string) (ref string, err error)
	RequestNetwork(netconf NetConfig, netviewName string) (network string, networks []string, err error)
	CreateGateway(cidrs []string, gw net.IP, netviewName string) (string, error)
	GetLockOnNetView(netViewName string) (*ibclient.NetworkViewLock, error)
	AllocateNewNetwork(netconf NetConfig, netView string) (string, error)
}

type InfobloxDriver struct {
	objMgr             ibclient.IBObjectManager
	Containers         []Container
	DefaultNetworkView string
	DefaultPrefixLen   uint
}

func (ibDrv *InfobloxDriver) RequestNetworkView(netviewName string) (string, error) {
	var netview *ibclient.NetworkView
	if netviewName == "" {
		netviewName = ibDrv.DefaultNetworkView
	}
	netview, _ = ibDrv.objMgr.GetNetworkView(netviewName)

	if netview == nil {
		netview, _ = ibDrv.objMgr.CreateNetworkView(netviewName)
	}

	log.Printf("RequestNetworkView: netview result is '%s'", *netview)
	return netview.Name, nil
}

func (ibDrv *InfobloxDriver) GetAddress(netviewName string, cidr string, ipAddr string, macAddr string) (*ibclient.FixedAddress, error) {
	if netviewName == "" {
		netviewName = ibDrv.DefaultNetworkView
	}
	fixedAddr, err := ibDrv.objMgr.GetFixedAddress(netviewName, cidr, ipAddr, macAddr)

	return fixedAddr, err
}

func (ibDrv *InfobloxDriver) RequestAddress(netviewName string, cidr string, ipAddr string, macAddr string, name string, vmID string) (string, error) {
	var fixedAddr *ibclient.FixedAddress
	var err error
	if netviewName == "" {
		netviewName = ibDrv.DefaultNetworkView
	}

	if len(macAddr) == 0 {
		log.Println("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.")
	} else {
		fixedAddr, err = ibDrv.objMgr.GetFixedAddress(netviewName, cidr, ipAddr, macAddr)
	}

	if fixedAddr == nil {
		fixedAddr, err = ibDrv.objMgr.AllocateIP(netviewName, cidr, ipAddr, macAddr, name, vmID)
	}

	log.Printf("RequestAddress: fixedAddr result is '%s'", *fixedAddr)
	return fmt.Sprintf("%s", fixedAddr.IPAddress), err
}

func (ibDrv *InfobloxDriver) UpdateAddress(fixedAddrRef string, macAddr string, name string, vmID string) (*ibclient.FixedAddress, error) {

	fixedAddr, err := ibDrv.objMgr.UpdateFixedAddress(fixedAddrRef, macAddr, name, vmID)
	if err != nil {
		log.Printf("UpdateAddress failed with error '%s'", err)
	}
	return fixedAddr, err
}

func (ibDrv *InfobloxDriver) ReleaseAddress(netviewName string, ipAddr string, macAddr string) (ref string, err error) {
	if netviewName == "" {
		netviewName = ibDrv.DefaultNetworkView
	}
	ref, err = ibDrv.objMgr.ReleaseIP(netviewName, "", ipAddr, macAddr)
	if ref == "" {
		log.Printf("ReleaseAddress: ***** IP Cannot be deleted '%s', '%s', '%s'! *******", netviewName, ipAddr, macAddr)
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
	for i := range ibDrv.Containers {
		if !ibDrv.Containers[i].exhausted {
			return &ibDrv.Containers[i]
		}
	}

	return nil
}

func (ibDrv *InfobloxDriver) resetContainers() {
	for i := range ibDrv.Containers {
		ibDrv.Containers[i].exhausted = false
		ibDrv.Containers[i].ContainerObj = nil
		ibDrv.Containers[i].NetworkView = ""
	}
}

func (ibDrv *InfobloxDriver) allocateNetworkHelper(netview string, prefixLen uint, name string) (network *ibclient.Network, err error) {
	container := ibDrv.nextAvailableContainer()
	for container != nil {
		log.Printf("Allocating network from Container:'%s'", container.NetworkContainer, container.exhausted)
		if container.ContainerObj == nil || container.NetworkView != netview {
			var err error
			container.ContainerObj, err = ibDrv.createNetworkContainer(netview, container.NetworkContainer)
			container.NetworkView = netview
			if err != nil || container.ContainerObj == nil {
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

func (ibDrv *InfobloxDriver) allocateNetwork(prefixLen uint, name string, netviewName string) (network *ibclient.Network, err error) {
	log.Printf("allocateNetwork: prefixLen='%d', name='%s'", prefixLen, name)
	if prefixLen == 0 {
		prefixLen = ibDrv.DefaultPrefixLen
	}
	network, err = ibDrv.allocateNetworkHelper(netviewName, prefixLen, name)
	if network == nil {
		ibDrv.resetContainers()
		network, err = ibDrv.allocateNetworkHelper(netviewName, prefixLen, name)
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
		if n, ok := network.Ea["Network Name"]; !ok || n != name {
			log.Printf("requestSpecificNetwork: network is already used '%s'", *network)
			return nil, nil
		}
	} else {
		networkByName, err := ibDrv.objMgr.GetNetwork(netview, "", ibclient.EA{"Network Name": name})
		if err != nil {
			return nil, err
		}
		if networkByName != nil {
			if networkByName.Cidr != subnet {
				log.Printf("requestSpecificNetwork: network name has different Cidr '%s'", networkByName.Cidr)
				return nil, nil
			}
		}
	}

	if network == nil {
		network, err = ibDrv.objMgr.CreateNetwork(netview, subnet, name)
		log.Printf("requestSpecificNetwork: CreateNetwork returns '%s', err='%s'", *network, err)
	}

	return network, err
}

func (ibDrv *InfobloxDriver) RequestNetwork(netconf NetConfig, netviewName string) (network string, networks []string, err error) {
	networks = make([]string, 100)
	var ibNetwork *ibclient.Network
	var ibNetworks *[]ibclient.Network
	// netviewName := netconf.IPAM.NetworkView
	cidr := net.IPNet{}
	log.Printf("RequestNetwork: IPAM.Subnet='%s'", netconf.IPAM.Subnet)
	if netconf.IPAM.Subnet.IP != nil {
		cidr = net.IPNet{IP: netconf.IPAM.Subnet.IP, Mask: netconf.IPAM.Subnet.Mask}
		ibNetwork, err = ibDrv.requestSpecificNetwork(netviewName, cidr.String(), netconf.Name)
	} else {
		//networkByName, err := ibDrv.objMgr.GetNetwork(netviewName, "", ibclient.EA{"Network Name": netconf.Name})
		networkByName, err := ibDrv.objMgr.GetNetworks(netviewName, "", ibclient.EA{"Network Name": netconf.Name})
		if err != nil {
			return "", nil, err
		}
		if networkByName != nil {
			ibNetworks = networkByName
			for i, network := range *ibNetworks {
				networks[i] = network.Cidr
			}
		} else {
			prefixLen := ibDrv.DefaultPrefixLen
			if netconf.IPAM.PrefixLength != 0 {
				prefixLen = netconf.IPAM.PrefixLength
			}
			ibNetwork, err = ibDrv.allocateNetwork(prefixLen, netconf.Name, netviewName)
		}
	}

	log.Printf("RequestNetwork: result='%s'", ibNetwork)
	if ibNetwork != nil {
		network = ibNetwork.Cidr
	}
	return network, networks, err
}

func (ibDrv *InfobloxDriver) CreateGateway(cidrs []string, gw net.IP, netviewName string) (string, error) {
	cidr := cidrs[0]
	gw = gw.To4() //making sure it is only 4 bytes
	//check for the format of gateway is in 0.0.0.x given by customer
	//it happens when no subnet given in the conf file
	if gw[0] == 0 {
		subnetIp, subnet, _ := net.ParseCIDR(cidr)
		subnetIp = subnetIp.To4()
		for index := 0; index <= 3; index++ {
			if gw[index] == 0 {
				gw[index] = subnetIp[index]
			}
		}
		if subnet.Contains(gw) == false {
			return "", fmt.Errorf("gateway given is invalid, should lie on subnet:'%s'", subnet)

		}
	}
	gateway := gw.String()
	//checking for gw ip already created ,if not creating
	gatewayIp, err := ibDrv.objMgr.GetFixedAddress(netviewName, cidr, gateway, "")
	if err == nil && gatewayIp != nil {
		log.Println("The Gateway already created")
	} else if gatewayIp == nil {
		gatewayIp, err = ibDrv.objMgr.AllocateIP(netviewName, cidr, gateway, "", "", "")
		if err != nil {
			log.Printf("Gateway creation failed with error:'%s'", err)
		}
	}
	return fmt.Sprintf("%s", gatewayIp), nil
}

func makeContainers(containerList string) []Container {
	var containers []Container

	parts := strings.Split(containerList, ",")
	for _, p := range parts {
		containers = append(containers, Container{p, "", nil, false})
	}

	return containers
}

func NewInfobloxDriver(objMgr ibclient.IBObjectManager, networkView string, networkContainer string, prefixLength uint) *InfobloxDriver {
	return &InfobloxDriver{
		objMgr:             objMgr,
		DefaultNetworkView: networkView,
		DefaultPrefixLen:   prefixLength,
		Containers:         makeContainers(networkContainer),
	}
}

// Gets the lock on network view
func (ibDrv *InfobloxDriver) GetLockOnNetView(netViewName string) (*ibclient.NetworkViewLock, error) {
	l := &ibclient.NetworkViewLock{Name: netViewName, ObjMgr: ibDrv.objMgr.(*ibclient.ObjectManager), LockEA: EA_PLUGIN_LOCK,
		LockTimeoutEA: EA_PLUGIN_LOCK_TIME}
	err := l.Lock()
	if err != nil {
		log.Printf("Error while getting lock on Network View %s: %s", netViewName, err)
		return l, fmt.Errorf("Error while getting lock on Network View %s: %s", netViewName, err)
	}
	return l, nil
}

//Allocates new network
func (ibDrv *InfobloxDriver) AllocateNewNetwork(netconf NetConfig, netView string) (subnet string, err error) {
	prefixLen := ibDrv.DefaultPrefixLen
	if netconf.IPAM.PrefixLength != 0 {
		prefixLen = netconf.IPAM.PrefixLength
	}
	ibNetwork, err := ibDrv.allocateNetwork(prefixLen, netconf.Name, netView)
	if ibNetwork != nil {
		subnet = ibNetwork.Cidr
	}
	return
}
