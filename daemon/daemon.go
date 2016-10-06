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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"runtime"

	"github.com/containernetworking/cni/pkg/types"
	. "github.com/infobloxopen/cni-infoblox"
	ibclient "github.com/infobloxopen/infoblox-go-client"
)

type Infoblox struct {
	Drv IBInfobloxDriver
}

func newInfoblox(drv IBInfobloxDriver) *Infoblox {
	return &Infoblox{
		Drv: drv,
	}
}

// Allocate acquires an IP from Infoblox for a specified container.
func (ib *Infoblox) Allocate(args *ExtCmdArgs, result *types.Result) (err error) {
	conf := NetConfig{}
	if err = json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("error parsing netconf: %v", err)
	}

	cidr := net.IPNet{IP: conf.IPAM.Subnet.IP, Mask: conf.IPAM.Subnet.Mask}
	netviewName := conf.IPAM.NetworkView
	log.Printf("RequestNetwork: '%s', '%s'", netviewName, cidr.String())
	netview, _ := ib.Drv.RequestNetworkView(netviewName)
	if netview == "" {
		return nil
	}

	subnet, _ := ib.Drv.RequestNetwork(conf)
	if subnet == "" {
		return nil
	}

	mac := args.IfMac

	log.Printf("RequestAddress: '%s', '%s', '%s'", netviewName, subnet, mac)
	ip, _ := ib.Drv.RequestAddress(netviewName, subnet, "", mac, args.ContainerID)

	ipn, _ := types.ParseCIDR(subnet)
	ipn.IP = net.ParseIP(ip)
	result.IP4 = &types.IPConfig{
		IP:      *ipn,
		Gateway: conf.IPAM.Gateway,
		Routes:  conf.IPAM.Routes,
	}

	log.Printf("Allocate result: '%s'", result)
	return nil
}

func (ib *Infoblox) Release(args *ExtCmdArgs, reply *struct{}) error {
	conf := NetConfig{}
	log.Printf("Release: called with args '%s'", *args)
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("error parsing netconf: %v", err)
	}

	ref, err := ib.Drv.ReleaseAddress(conf.IPAM.NetworkView, "", args.IfMac)
	log.Printf("Fixed Address released: '%s'", ref)

	return err
}

func getListener(driverSocket *DriverSocket) (net.Listener, error) {
	socketFile := driverSocket.SetupSocket()

	return net.Listen("unix", socketFile)
}

func getInfobloxDriver(config *Config) *InfobloxDriver {
	hostConfig := ibclient.HostConfig{
		Host:     config.GridHost,
		Version:  config.WapiVer,
		Port:     config.WapiPort,
		Username: config.WapiUsername,
		Password: config.WapiPassword,
	}
	transportConfig := ibclient.NewTransportConfig(
		config.SslVerify,
		config.HttpRequestTimeout,
		config.HttpPoolConnections,
	)

	requestBuilder := &ibclient.WapiRequestBuilder{}
	requestor := &ibclient.WapiHttpRequestor{}
	conn, _ := ibclient.NewConnector(hostConfig, transportConfig,
		requestBuilder, requestor)

	objMgr := ibclient.NewObjectManager(conn, "Rkt", "RktEngineID")

	return NewInfobloxDriver(objMgr, config.NetworkView, config.NetworkContainer, config.PrefixLength)
}

func runDaemon(config *Config) {
	// since other goroutines (on separate threads) will change namespaces,
	// ensure the RPC server does not get scheduled onto those
	runtime.LockOSThread()

	log.Printf("Config is '%v'\n", *config)

	driverSocket := NewDriverSocket(config.SocketDir, config.DriverName)
	l, err := getListener(driverSocket)

	if err != nil {
		log.Printf("Error getting listener: %v", err)
		return
	}

	ibDrv := getInfobloxDriver(config)

	ib := newInfoblox(ibDrv)
	rpc.Register(ib)
	rpc.HandleHTTP()
	http.Serve(l, nil)
}

func main() {
	config := LoadConfig()
	runDaemon(config)
}
