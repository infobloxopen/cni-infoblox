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
	"strings"
	"os"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/utils/hwaddr"
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
func (ib *Infoblox) Allocate(args *ExtCmdArgs, result *current.Result) (err error) {
	conf := NetConfig{}

	log.Printf("Allocate: called with args '%s'", *args)
	/* Sample args passed in K8s
		ContainerID: 85f177f2f1981087309589281979e1190931a9f3d7840660ac8dd9eaeb5685fb
		Netns       /proc/2617/ns/net
		IfName      eth0
		Args        IgnoreUnknown=1;K8S_POD_NAMESPACE=default;K8S_POD_NAME=test-infoblox-deployment-8478849b97-p2jhp;K8S_POD_INFRA_CONTAINER_ID=<>
		Path        /opt/macvlan/bin:/opt/cni/bin
		StdinData   {"cniVersion":"","ipam":{"gateway":"10.0.0.1","network-view":"cni_view","subnet":"10.0.0.0/24","type":"infoblox"},"master":"eth1","name":"ipam-test","type":"macvlan"}}
	66:c2:1c:94:6e:e5}
	*/

	if err = json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("error parsing netconf: %v", err)
	}

	cidr := net.IPNet{IP: conf.IPAM.Subnet.IP, Mask: conf.IPAM.Subnet.Mask}
	netviewName := conf.IPAM.NetworkView
	netCreateCheck := os.Getenv("CHECK_NETWORK")
	// Create network only if CHECK_NETWORK environment varible is not defined or set as true
	subnet := cidr.String()
	log.Printf("RequestNetwork: '%s', '%s'", netviewName, cidr.String())
	if len(netCreateCheck) == 0 ||  os.Getenv("CHECK_NETWORK") == "true" {
			gw := conf.IPAM.Gateway
			netview, _ := ib.Drv.RequestNetworkView(netviewName)
			if netview == "" {
				return nil
			}
			subnet, _ := ib.Drv.RequestNetwork(conf, netview)
			if subnet == "" {
				return nil
			}

			//cni is not calling gateway creation call, so it is implemented here
			//if gateway is not provided in net conf file by customer, it wont create as for now
			if gw != nil {
				if _, err := ib.Drv.CreateGateway(subnet, gw, netviewName); err != nil {
					return fmt.Errorf("error creating gateway:%v", err)
				}
			}
	}
	mac := args.IfMac

	return ib.requestAddress(conf, args, result, netviewName, subnet, mac)
}

func (ib *Infoblox) requestAddress(conf NetConfig, args *ExtCmdArgs, result *current.Result, netviewName string, cidr string, macAddr string) (err error) {

	// In Kubernetes to get the container name/hostname
	containerName := ""
	str1 := strings.Split(args.Args, "K8S_POD_NAME=")
	if len(str1) != 1 {
		str2 := strings.Split(str1[1], ";")
		containerName = str2[0]
	}

	log.Printf("RequestAddress: '%s', '%s', '%s'", netviewName, cidr, macAddr)
	ip, _ := ib.Drv.RequestAddress(netviewName, cidr, "", macAddr, containerName, args.ContainerID)

	log.Printf("Allocated IP: '%s'", ip)

	// As bridge plugin in CNI generates MAC address based on ip, so the daemon also generating MAC address based on
	// ip and updating GRID host with the new MAC address
	if conf.Type == "bridge" {
		hwAddr, err := hwaddr.GenerateHardwareAddr4(net.ParseIP(ip), hwaddr.PrivateMACPrefix)
		if err != nil {
			log.Printf("Problem while generating hardware address using ip: %s", err)
			return err
		}

		err = ib.updateAddress(netviewName, cidr, ip, hwAddr.String(), containerName)
		if err != nil {
			log.Printf("Problem while updating MacAddress: %s", err)
			return err
		}
	}
	ipn, _ := types.ParseCIDR(cidr)
	ipn.IP = net.ParseIP(ip)
	ipConfig := &current.IPConfig{
		Version: "4",
		Address: *ipn,
		Gateway: conf.IPAM.Gateway,
	}
	routes := convertRoutesToCurrent(conf.IPAM.Routes)
	result.IPs = []*current.IPConfig{ipConfig}
	result.Routes = routes

	log.Printf("Allocate result: '%s'", result)
	return nil
}

func (ib *Infoblox) updateAddress(netviewName string, cidr string, ipAddr string, macAddr string, name string) error {

	fixedAddr, err := ib.Drv.GetAddress(netviewName, cidr, ipAddr, "")
	if err != nil {
		return err
	}
	updatedFixedAddr, err := ib.Drv.UpdateAddress(fixedAddr.Ref, macAddr, name, "")
	if err != nil {
		return err
	}
	log.Printf("UpdatedAddress: fixedAddr result is '%s'", *updatedFixedAddr)
	return nil
}

func convertRoutesToCurrent(routes []types.Route) []*types.Route {
	var currentRoutes []*types.Route
	for _, r := range routes {
		currentRoutes = append(currentRoutes, &types.Route{
			Dst: r.Dst,
			GW:  r.GW,
		})
	}
	return currentRoutes
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

	objMgr := ibclient.NewObjectManager(conn, "Kubernetes", config.ClusterName)
	CheckForCloudLicense(objMgr)
	return NewInfobloxDriver(objMgr, config.NetworkView, config.NetworkContainer, config.PrefixLength)
}

func runDaemon(config *Config) {
	// since other goroutines (on separate threads) will change namespaces,
	// ensure the RPC server does not get scheduled onto those
	runtime.LockOSThread()

	configToLog := *config
	configToLog.WapiPassword = "******"
	log.Printf("Config is '%v'\n", configToLog)

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
