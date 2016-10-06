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
	ContainerObj     *ibclient.NetworkContainer
	exhausted        bool
}

type IBInfobloxDriver interface {
	RequestNetworkView(netviewName string) (string, error)
	RequestAddress(netviewName string, cidr string, ipAddr string, macAddr string, vmID string) (string, error)
	ReleaseAddress(netviewName string, ipAddr string, macAddr string) (ref string, err error)
	RequestNetwork(netconf NetConfig) (network string, err error)
}

type InfobloxDriver struct {
	objMgr     ibclient.IBObjectManager
	Containers []Container

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

func (ibDrv *InfobloxDriver) RequestAddress(netviewName string, cidr string, ipAddr string, macAddr string, vmID string) (string, error) {
	var fixedAddr *ibclient.FixedAddress

	if netviewName == "" {
		netviewName = ibDrv.DefaultNetworkView
	}

	if len(macAddr) == 0 {
		log.Println("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.")
	} else {
		fixedAddr, _ = ibDrv.objMgr.GetFixedAddress(netviewName, cidr, ipAddr, macAddr)
	}

	if fixedAddr == nil {
		fixedAddr, _ = ibDrv.objMgr.AllocateIP(netviewName, cidr, ipAddr, macAddr, vmID)
	}

	log.Printf("RequestAddress: fixedAddr result is '%s'", *fixedAddr)
	return fmt.Sprintf("%s", fixedAddr.IPAddress), nil
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
	}
}

func (ibDrv *InfobloxDriver) allocateNetworkHelper(netview string, prefixLen uint, name string) (network *ibclient.Network, err error) {
	log.Printf("allocateNetworkHelper: netview='%s', prefixLen='%d', name='%s'", netview, prefixLen, name)
	container := ibDrv.nextAvailableContainer()
	for container != nil {
		log.Printf("Allocating network from Container:'%s'", container.NetworkContainer)
		if container.ContainerObj == nil {
			var err error
			container.ContainerObj, err = ibDrv.createNetworkContainer(netview, container.NetworkContainer)
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

func (ibDrv *InfobloxDriver) allocateNetwork(prefixLen uint, name string) (network *ibclient.Network, err error) {
	log.Printf("allocateNetwork: prefixLen='%d', name='%s'", prefixLen, name)
	if prefixLen == 0 {
		prefixLen = ibDrv.DefaultPrefixLen
	}
	network, err = ibDrv.allocateNetworkHelper(ibDrv.DefaultNetworkView, prefixLen, name)
	if network == nil {
		ibDrv.resetContainers()
		network, err = ibDrv.allocateNetworkHelper(ibDrv.DefaultNetworkView, prefixLen, name)
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

func (ibDrv *InfobloxDriver) RequestNetwork(netconf NetConfig) (network string, err error) {
	var ibNetwork *ibclient.Network
	netviewName := netconf.IPAM.NetworkView
	cidr := net.IPNet{}
	log.Printf("RequestNetwork: IPAM.Subnet='%s'", netconf.IPAM.Subnet)
	if netconf.IPAM.Subnet.IP != nil {
		cidr = net.IPNet{IP: netconf.IPAM.Subnet.IP, Mask: netconf.IPAM.Subnet.Mask}
		ibNetwork, err = ibDrv.requestSpecificNetwork(netviewName, cidr.String(), netconf.Name)
	} else {
		networkByName, err := ibDrv.objMgr.GetNetwork(netviewName, "", ibclient.EA{"Network Name": netconf.Name})
		if err != nil {
			return "", err
		}
		if networkByName != nil {
			log.Printf("RequestNetwork: GetNetwork by name returns '%s'", *networkByName)
			ibNetwork = networkByName
		} else {
			if netviewName == "" {
				netviewName = ibDrv.DefaultNetworkView
			}
			if netviewName == ibDrv.DefaultNetworkView {
				prefixLen := ibDrv.DefaultPrefixLen
				if netconf.IPAM.PrefixLength != 0 {
					prefixLen = netconf.IPAM.PrefixLength
				}
				ibNetwork, err = ibDrv.allocateNetwork(prefixLen, netconf.Name)
			} else {
				log.Printf("RequestNetwork: Incorrect Network View name specified='%s'", netviewName)
				return "", nil
			}
		}
	}

	log.Printf("RequestNetwork: result='%s'", ibNetwork)
	if ibNetwork != nil {
		network = ibNetwork.Cidr
	}
	return network, err
}

func makeContainers(containerList string) []Container {
	var containers []Container

	parts := strings.Split(containerList, ",")
	for _, p := range parts {
		containers = append(containers, Container{p, nil, false})
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
