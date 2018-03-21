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
	"flag"
	"net"

	"github.com/containernetworking/cni/pkg/types"
)

const (
	HTTP_REQUEST_TIMEOUT  = 60
	HTTP_POOL_CONNECTIONS = 10
)

type GridConfig struct {
	GridHost            string
	WapiVer             string
	WapiPort            string
	WapiUsername        string
	WapiPassword        string
	SslVerify           string
	HttpRequestTimeout  int
	HttpPoolConnections int
	HttpPoolMaxSize     int
}

type DriverConfig struct {
	SocketDir        string
	DriverName       string
	NetworkView      string
	NetworkContainer string
	PrefixLength     uint
}

type Config struct {
	GridConfig
	DriverConfig
}

func LoadConfig() (config *Config) {
	config = new(Config)

	flag.StringVar(&config.GridHost, "grid-host", "192.168.124.200", "IP of Infoblox Grid Host")
	flag.StringVar(&config.WapiVer, "wapi-version", "2.5", "Infoblox WAPI Version.")
	flag.StringVar(&config.WapiPort, "wapi-port", "443", "Infoblox WAPI Port.")
	flag.StringVar(&config.WapiUsername, "wapi-username", "", "Infoblox WAPI Username")
	flag.StringVar(&config.WapiPassword, "wapi-password", "", "Infoblox WAPI Password")
	flag.StringVar(&config.SslVerify, "ssl-verify", "false", "Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate.")
	flag.StringVar(&config.NetworkView, "network-view", "default", "Infoblox Network View")
	flag.StringVar(&config.NetworkContainer, "network-container", "172.18.0.0/16", "Subnets will be allocated from this container if subnet is not specified in network config file")
	flag.UintVar(&config.PrefixLength, "prefix-length", 24, "The CIDR prefix length when allocating a subnet from Network Container")
	config.HttpRequestTimeout = HTTP_REQUEST_TIMEOUT
	config.HttpPoolConnections = HTTP_POOL_CONNECTIONS

	flag.StringVar(&config.SocketDir, "socket-dir", GetDefaultSocketDir(), "Directory where Infoblox IPAM daemon sockets are created")
	flag.StringVar(&config.DriverName, "driver-name", "infoblox", "Name of Infoblox IPAM driver")

	flag.Parse()

	return config
}

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
