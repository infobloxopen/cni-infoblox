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
	"net"
	"net/rpc"
	"path/filepath"
	"sync"

	"github.com/containernetworking/cni/pkg/ns"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
)

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.Legacy)
}

type InterfaceInfo struct {
	iface *net.Interface
	wg    sync.WaitGroup
}

func getMacAddress(netns string, ifaceName string) (mac string) {
	var err error
	ifaceInfo := &InterfaceInfo{}
	if netns == "" {
		ifaceInfo.iface, err = net.InterfaceByName(ifaceName)
		mac = ifaceInfo.iface.HardwareAddr.String()

	} else {
		errCh := make(chan error, 1)
		ifaceInfo.wg.Add(1)
		go func() {
			errCh <- ns.WithNetNSPath(netns, func(_ ns.NetNS) error {
				defer ifaceInfo.wg.Done()

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

func cmdAdd(args *skel.CmdArgs) error {
	result := types.Result{}
	extArgs := &ExtCmdArgs{CmdArgs: *args}

	mac := getMacAddress(args.Netns, args.IfName)

	extArgs.IfMac = mac
	if err := rpcCall("Infoblox.Allocate", extArgs, &result); err != nil {
		return err
	}

	return result.Print()
}

func cmdDel(args *skel.CmdArgs) error {
	result := struct{}{}
	extArgs := &ExtCmdArgs{CmdArgs: *args}

	mac := getMacAddress(args.Netns, args.IfName)
	extArgs.IfMac = mac
	if err := rpcCall("Infoblox.Release", extArgs, &result); err != nil {
		return fmt.Errorf("error dialing Infoblox daemon: %v", err)
	}
	return nil
}

func rpcCall(method string, args *ExtCmdArgs, result interface{}) error {
	conf := NetConfig{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("error parsing netconf: %v", err)
	}

	client, err := rpc.DialHTTP("unix", NewDriverSocket(conf.IPAM.SocketDir, conf.IPAM.Type).GetSocketFile())
	if err != nil {
		return fmt.Errorf("error dialing Infoblox daemon: %v", err)
	}

	// The daemon may be running under a different working dir
	// so make sure the netns path is absolute.
	netns, err := filepath.Abs(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to make %q an absolute path: %v", args.Netns, err)
	}
	args.Netns = netns

	err = client.Call(method, args, result)
	if err != nil {
		return fmt.Errorf("error calling %v: %v", method, err)
	}

	return nil
}
