// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	//	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	//	"os"
	"runtime"
	"sync"

	"github.com/containernetworking/cni/pkg/ns"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	ibclient "github.com/infobloxopen/infoblox-go-client"
)

type IPAMConfig struct {
	Type             string        `json:"type"`
	SocketDir        string        `json:"socket-dir"`
	NetworkView      string        `json:"network-view"`
	NetworkContainer string        `json:"network-container"`
	PrefixLength     uint          `json:"prefix-length"`
	Subnet           types.IPNet   `json:"subnet"`
	Gateway          net.IP        `json:"gateway"`
	Routes           []types.Route `json:"routes"`
}

type NetConfig struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Bridge    string      `json:"bridge"`
	IsGateway bool        `json:"isGateway"`
	IPAM      *IPAMConfig `json:"ipam"`
}

type Infoblox struct {
	//	mux    sync.Mutex
	//	leases map[string]*DHCPLease
	Drv *InfobloxDriver
}

func newInfoblox(drv *InfobloxDriver) *Infoblox {
	return &Infoblox{
		Drv: drv,
	}
}

func listInterfaces(netns string) {
	fmt.Printf("Listing interfaces in namespace '%s'\n", netns)

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting interfaces: '%s'\n", err)
		return
	}

	for i, iface := range ifaces {
		fmt.Printf("interface %d: '%s'\n", i, iface.Name)
	}
}

type InterfaceInfo struct {
	iface *net.Interface
	wg    sync.WaitGroup
}

func getMacAddress(netns string, ifaceName string) (mac string) {
	var err error
	ifaceInfo := &InterfaceInfo{}
	if netns == "" {
		listInterfaces(netns)
		fmt.Printf("in getMacAddress(1), ifaceName is '%s'\n", ifaceName)
		ifaceInfo.iface, err = net.InterfaceByName(ifaceName)
		mac = ifaceInfo.iface.HardwareAddr.String()

	} else {
		errCh := make(chan error, 1)
		ifaceInfo.wg.Add(1)
		go func() {
			errCh <- ns.WithNetNSPath(netns, func(_ ns.NetNS) error {
				defer ifaceInfo.wg.Done()

				listInterfaces(netns)
				fmt.Printf("in getMacAddress(2), ifaceName is '%s'\n", ifaceName)
				ifaceInfo.iface, err = net.InterfaceByName(ifaceName)
				if err != nil {
					return fmt.Errorf("error looking up interface '%s': '%s'", ifaceName, err)
				}
				return nil
			})
		}()

		if err = <-errCh; err != nil {
			fmt.Printf("%s\n", err)
		} else {
			mac = ifaceInfo.iface.HardwareAddr.String()
		}
	}

	return mac
}

// Allocate acquires an IP from Infoblox for a specified container.
func (ib *Infoblox) Allocate(args *skel.CmdArgs, result *types.Result) (err error) {
	conf := NetConfig{}
	if err = json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("error parsing netconf: %v", err)
	}

	cidr := net.IPNet{IP: conf.IPAM.Subnet.IP, Mask: conf.IPAM.Subnet.Mask}
	netviewName := conf.IPAM.NetworkView
	if netviewName == "" {
		netviewName = ib.Drv.networkView
	}
	log.Printf("RequestNetwork: '%s', '%s'\n", netviewName, cidr.String())
	subnet, _ := ib.Drv.RequestNetwork(conf)
	if subnet == "" {
		return nil
	}

	gw := ""
	log.Printf("RequestNetwork: conf.Type is '%s', conf.IsGateway is '%s, conf.Bridge '%s'\n", conf.Type, conf.IsGateway, conf.Bridge)
	if conf.Type == "bridge" && conf.IsGateway && conf.Bridge != "" {
		gwMac := getMacAddress("", conf.Bridge)
		log.Printf("RequestNetwork: gwMac is '%s'\n", gwMac)

		if gwMac != "" {

			gwIP := ""
			if conf.IPAM.Gateway != nil {
				gwIP = conf.IPAM.Gateway.String()
			}

			gw, _ = ib.Drv.RequestAddress(netviewName, subnet, gwIP, gwMac, "")
			log.Printf("RequestNetwork: gwIP is '%s', gw is '%s'\n", gwIP, gw)
		}
	}

	mac := getMacAddress(args.Netns, args.IfName)

	fmt.Printf("RequestAddress: '%s', '%s', '%s'\n", netviewName, subnet, mac)
	ip, _ := ib.Drv.RequestAddress(netviewName, subnet, "", mac, args.ContainerID)

	//fmt.Printf("In Allocate(), args: '%s'\n", args)
	//fmt.Printf("In Allocate(), conf: '%s'\n", conf)

	ipn, _ := types.ParseCIDR(subnet)
	ipn.IP = net.ParseIP(ip)
	fmt.Printf("ip: '%s'\n", ip)
	fmt.Printf("ipn: '%s'\n", *ipn)
	result.IP4 = &types.IPConfig{
		IP: *ipn,
		//		Gateway: conf.IPAM.Gateway,
		Gateway: net.ParseIP(gw),
		Routes:  conf.IPAM.Routes,
	}

	return nil
}

// Release stops maintenance of the lease acquired in Allocate()
// and sends a release msg to the DHCP server.
func (ib *Infoblox) Release(args *skel.CmdArgs, reply *struct{}) error {
	conf := NetConfig{}
	fmt.Printf("Infoblox.Release called, args '%v'\n", *args)
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("error parsing netconf: %v", err)
	}

	mac := getMacAddress(args.Netns, args.IfName)
	fmt.Printf("Infoblox.Release called, mac is '%s'\n", mac)

	ref, err := ib.Drv.ReleaseAddress(conf.IPAM.NetworkView, "", mac)
	fmt.Printf("IP Address released: '%s'\n", ref)

	return err
}

func getListener(driverSocket *DriverSocket) (net.Listener, error) {
	socketFile := driverSocket.SetupSocket()

	return net.Listen("unix", socketFile)
}

func runDaemon(config *Config) {
	// since other goroutines (on separate threads) will change namespaces,
	// ensure the RPC server does not get scheduled onto those
	runtime.LockOSThread()

	fmt.Printf("Config is '%s'\n", config)

	conn, err := ibclient.NewConnector(
		config.GridHost,
		config.WapiVer,
		config.WapiPort,
		config.WapiUsername,
		config.WapiPassword,
		config.SslVerify,
		config.HttpRequestTimeout,
		config.HttpPoolConnections,
		config.HttpPoolMaxSize)

	log.Printf("Socket Dir: '%s'\n", config.SocketDir)
	log.Printf("Driver Name: '%s'\n", config.DriverName)
	driverSocket := NewDriverSocket(config.SocketDir, config.DriverName)
	l, err := getListener(driverSocket)

	objMgr := ibclient.NewObjectManager(conn, "Rkt", "RktEngineID")

	ibDrv := NewInfobloxDriver(objMgr, config.NetworkView, config.NetworkContainer, config.PrefixLength)

	if err != nil {
		log.Printf("Error getting listener: %v", err)
		return
	}

	ib := newInfoblox(ibDrv)
	rpc.Register(ib)
	rpc.HandleHTTP()
	http.Serve(l, nil)
}
