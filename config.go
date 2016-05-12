package main

import (
	"flag"
	"fmt"
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
	PluginDir              string
	DriverName             string
}

type Config struct {
	Daemon       bool
	GridConfig
	DriverConfig
}

func LoadConfig() (config *Config) {
	config = new(Config)
	fmt.Printf("Args are: '%s'\n", flag.Args())

	flag.BoolVar(&config.Daemon, "daemon", false, "Daemon Mode")
	flag.StringVar(&config.GridHost, "grid-host", "192.168.124.200", "IP of Infoblox Grid Host")
	flag.StringVar(&config.WapiVer, "wapi-version", "2.0", "Infoblox WAPI Version.")
	flag.StringVar(&config.WapiPort, "wapi-port", "443", "Infoblox WAPI Port.")
	flag.StringVar(&config.WapiUsername, "wapi-username", "", "Infoblox WAPI Username")
	flag.StringVar(&config.WapiPassword, "wapi-password", "", "Infoblox WAPI Password")
	flag.StringVar(&config.SslVerify, "ssl-verify", "false", "Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate.")
	config.HttpRequestTimeout = HTTP_REQUEST_TIMEOUT
	config.HttpPoolConnections = HTTP_POOL_CONNECTIONS
	config.HttpPoolMaxSize = HTTP_POOL_MAX_SIZE

	flag.StringVar(&config.PluginDir, "plugin-dir", "/run/cni", "Docker plugin directory where driver socket is created")
	flag.StringVar(&config.DriverName, "driver-name", "infoblox", "Name of Infoblox IPAM driver")

	flag.Parse()
	fmt.Printf("Args are: '%s'\n", flag.Args())

	return config
}
