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
	"os"
	"runtime"
	"sync"

	"github.com/containernetworking/cni/pkg/ns"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	ibclient "github.com/infobloxopen/infoblox-go-client"
)

type IPAMConfig struct {
	Type             string        `json:"type"`
	NetworkView      string        `json:"network-view"`
	NetworkContainer string        `json:"network-container"`
	PrefixLength     uint          `json:"prefix-length"`
	Subnet           types.IPNet   `json:"subnet"`
	Gateway          net.IP        `json:"gateway"`
	Routes           []types.Route `json:"routes"`
}

type NetConfig struct {
	Name string      `json:"name"`
	IPAM *IPAMConfig `json:"ipam"`
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

type InterfaceInfo struct {
	iface *net.Interface
	wg    sync.WaitGroup
}


func getMacAddress(args *skel.CmdArgs) (mac string) {
	var err error
	ifaceInfo := &InterfaceInfo{}
	errCh := make(chan error, 1)
	ifaceInfo.wg.Add(1)
	go func() {
		errCh <- ns.WithNetNSPath(args.Netns, true, func(_ *os.File) error {
			defer ifaceInfo.wg.Done()

			ifaceInfo.iface, err = net.InterfaceByName(args.IfName)
			if err != nil {
				return fmt.Errorf("error looking up interface '%s': '%s'", args.IfName, err)
			}
			return nil
		})
	}()

	if err = <-errCh; err != nil {
		fmt.Printf("%s\n", err)
	} else {
		mac = ifaceInfo.iface.HardwareAddr.String()
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
	fmt.Printf("RequestNetwork: '%s', '%s'\n", conf.IPAM.NetworkView, cidr.String())
	netviewName := conf.IPAM.NetworkView
	if netviewName == "" {
		netviewName = ib.Drv.networkView
	}
	subnet, gw, _ := ib.Drv.RequestNetwork(conf)

	mac := getMacAddress(args)

	fmt.Printf("RequestAddress: '%s', '%s', '%s'\n", netviewName, subnet, mac)
	ip, _ := ib.Drv.RequestAddress(netviewName, subnet, mac, args.ContainerID)

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

	mac := getMacAddress(args)
	fmt.Printf("Infoblox.Release called, mac is '%s'\n", mac)

	//	ref, _ := ib.Drv.ReleaseIP
	return nil
	/*
		if l := d.getLease(args.ContainerID, conf.Name); l != nil {
			l.Stop()
			return nil
		}

		return fmt.Errorf("lease not found: %v/%v", args.ContainerID, conf.Name)
	*/
}

/*
func (d *DHCP) getLease(contID, netName string) *DHCPLease {
	d.mux.Lock()
	defer d.mux.Unlock()

	// TODO(eyakubovich): hash it to avoid collisions
	l, ok := d.leases[contID+netName]
	if !ok {
		return nil
	}
	return l
}

func (d *DHCP) setLease(contID, netName string, l *DHCPLease) {
	d.mux.Lock()
	defer d.mux.Unlock()

	// TODO(eyakubovich): hash it to avoid collisions
	d.leases[contID+netName] = l
}

func getListener() (net.Listener, error) {
	l, err := activation.Listeners(true)
	if err != nil {
		return nil, err
	}

	switch {
	case len(l) == 0:
		if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
			return nil, err
		}
		return net.Listen("unix", socketPath)

	case len(l) == 1:
		if l[0] == nil {
			return nil, fmt.Errorf("LISTEN_FDS=1 but no FD found")
		}
		return l[0], nil

	default:
		return nil, fmt.Errorf("Too many (%v) FDs passed through socket activation", len(l))
	}
}
*/

func dirExists(dirname string) (bool, error) {
	fileInfo, err := os.Stat(dirname)
	if err == nil {
		if fileInfo.IsDir() {
			return true, nil
		} else {
			return false, nil
		}
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createDir(dirname string) error {
	return os.MkdirAll(dirname, 0700)
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)

	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func deleteFile(filePath string) error {
	return os.Remove(filePath)
}

func setupSocket(pluginDir string, driverName string) string {
	exists, err := dirExists(pluginDir)
	if err != nil {
		log.Panicf("Stat Plugin Directory error '%s'", err)
		os.Exit(1)
	}
	if !exists {
		err = createDir(pluginDir)
		if err != nil {
			log.Panicf("Create Plugin Directory error: '%s'", err)
			os.Exit(1)
		}
		log.Printf("Created Plugin Directory: '%s'", pluginDir)
	}

	socketFile := pluginDir + "/" + driverName + ".sock"
	fmt.Printf("socketFile: '%s'\n", socketFile)
	exists, err = fileExists(socketFile)
	if err != nil {
		log.Panicf("Stat Socket File error: '%s'", err)
		os.Exit(1)
	}
	if exists {
		err = deleteFile(socketFile)
		if err != nil {
			log.Panicf("Delete Socket File error: '%s'", err)
			os.Exit(1)
		}
		log.Printf("Deleted Old Socket File: '%s'", socketFile)
	}

	return socketFile
}

func getListener(pluginDir string, driverName string) (net.Listener, error) {
	fmt.Printf("pluginDir: '%s'\n", pluginDir)
	fmt.Printf("driverName: '%s'\n", driverName)

	socketFile := setupSocket(pluginDir, driverName)

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

	l, err := getListener(config.PluginDir, config.DriverName)

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
