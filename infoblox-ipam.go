package main

import (
//	"errors"
	"fmt"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"log"
//	"strconv"
//	"strings"
)

type InfobloxDriver struct {
	objMgr              *ibclient.ObjectManager
}

func (ibDrv *InfobloxDriver) RequestAddress(netviewName string, cidr string, macAddr string, vmID string) (string,  error) {
	if len(macAddr) == 0 {
		log.Printf("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.\n")
	}
	fixedAddr, _ := ibDrv.objMgr.AllocateIP(netviewName, cidr, macAddr, vmID)

	fmt.Printf("RequestAddress: fixedAddr is '%s'\n", *fixedAddr)
	return fmt.Sprintf("%s", fixedAddr.IPAddress), nil
}

/*
func (ibDrv *InfobloxDriver) ReleaseAddress(netviewName string, ) (map[string]interface{}, error) {
	network := ibclient.BuildNetworkFromRef(v.PoolID)
	ref, _ := ibDrv.objMgr.ReleaseIP(network.NetviewName, v.Address)
	if ref == "" {
		log.Printf("***** IP Cannot be deleted '%s'! *******\n", v.Address)
	}

	return map[string]interface{}{}, nil
}
*/

func (ibDrv *InfobloxDriver) requestSpecificNetwork(netview string, subnet string) (*ibclient.Network, error) {
	network, err := ibDrv.objMgr.GetNetwork(netview, subnet)
	if network != nil {
		fmt.Printf("GetNetwork: network is '%s'\n", *network)
	}
	if network == nil {
		network, err = ibDrv.objMgr.CreateNetwork(netview, subnet)
		fmt.Printf("CreateNetwork: network is '%s', err is '%s'\n", *network, err)
	}

	return network, err
}

func (ibDrv *InfobloxDriver) RequestNetwork(netviewName string, cidr string) (res string, err error) {
	var network *ibclient.Network
	network, err = ibDrv.requestSpecificNetwork(netviewName, cidr)

	fmt.Printf("RequestNetwork: network is '%s'\n", *network)
	res = ""
	if network != nil {
		res = network.Cidr
	}

	return res, err
}

func NewInfobloxDriver(objMgr *ibclient.ObjectManager) *InfobloxDriver {
	return &InfobloxDriver{
		objMgr: objMgr,
	}
}
