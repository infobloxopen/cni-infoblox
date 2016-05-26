package main

import (
	"flag"
)

const (
	HTTP_REQUEST_TIMEOUT  = 120
	HTTP_POOL_CONNECTIONS = 100
	HTTP_POOL_MAX_SIZE    = 100
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
	Daemon bool
	GridConfig
	DriverConfig
}

func LoadConfig() (config *Config) {
	config = new(Config)

	flag.BoolVar(&config.Daemon, "daemon", false, "Daemon Mode")
	flag.StringVar(&config.GridHost, "grid-host", "192.168.124.200", "IP of Infoblox Grid Host")
	flag.StringVar(&config.WapiVer, "wapi-version", "2.0", "Infoblox WAPI Version.")
	flag.StringVar(&config.WapiPort, "wapi-port", "443", "Infoblox WAPI Port.")
	flag.StringVar(&config.WapiUsername, "wapi-username", "", "Infoblox WAPI Username")
	flag.StringVar(&config.WapiPassword, "wapi-password", "", "Infoblox WAPI Password")
	flag.StringVar(&config.SslVerify, "ssl-verify", "false", "Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate.")
	flag.StringVar(&config.NetworkView, "network-view", "default", "Infoblox Network View")
	flag.StringVar(&config.NetworkContainer, "network-container", "172.18.0.0/16", "Subnets will be allocated from this container if subnet is not specified in IPAM config")
	flag.UintVar(&config.PrefixLength, "prefix-length", 24, "The default CIDR prefix length when allocating a subnet from Network Container")
	config.HttpRequestTimeout = HTTP_REQUEST_TIMEOUT
	config.HttpPoolConnections = HTTP_POOL_CONNECTIONS
	config.HttpPoolMaxSize = HTTP_POOL_MAX_SIZE

	flag.StringVar(&config.SocketDir, "socket-dir", GetDefaultSocketDir(), "Directory where Infoblox IPAM daemon sockets are created")
	flag.StringVar(&config.DriverName, "driver-name", "infoblox", "Name of Infoblox IPAM driver")

	flag.Parse()

	return config
}
