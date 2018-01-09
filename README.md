CNI IPAM Driver
===============

Infoblox IPAM Driver for CNI
----------------------------

cni-infoblox is an IPAM driver for CNI that interfaces with Infoblox to provide IP Address Management
service. CNI is the generic plugin-based networking layer for supporting container runtime environments,
of which rkt is one.

For a detailed description of the driver, including a step by step deployment example, refer to the
"CNI Networking and IPAM" community blog on the Infolox website:
https://community.infoblox.com/t5/Community-Blog/CNI-Networking-and-IPAM/ba-p/7828

Prerequisite
------------
To use the plugin, you need access to the Infoblox DDI product. For evaluation purposes, you can download a
virtual version of the product from the Infoblox Download Center (https://www.infoblox.com/infoblox-download-center)
Alternatively, if you are an existing Infoblox customer, you can download it from the support site.

Refer to CONFIG.md for details on vNIOS configuration.

Build
-----
For dependencies and build instructions, refer to ```BUILD.md```.

CNI Configuration
-----------------
This section concerns only with CNI network configuration as it relates to the Infoblox IPAM Driver.
For details on CNI configuration in general, refer to https://github.com/containernetworking/cni/blob/master/README.md

To instruct CNI to execute the Infoblox IPAM plugin for a particular network, specify "infoblox" as the IPAM "type"
in the CNI network configuration file (netconf). CNI configuration files in a rkt environment is typically
localted in ```/etc/rkt/net.d```

For example (/etc/rkt/net.d/10-net-1.conf):

```
{
    "name": "net-1",
    "ipam": {
        "type": "infoblox",
        "subnet": "172.18.1.0/24",
		"gateway": "172.18.1.1",
		"routes": [
			{"dst": "172.18.0.0/24"}
		],
		"network-view": "priv-view"
    }
}
```

The following are the IPAM attributes:
- "type": specifies the plugin type and is also the file name of the plugin executable.
- "subnet": specifies the CIDR to be used for the network. This is a well-known CNI attribute and is used by the driver.
- "gateway": specifies the gateway for the network. This is a well-known CNI attribute and is simply passed through to CNI.
- "routes": specifies the routes for the network. This is a well-known CNI attribute and is simply passed through to CNI.
- "network-view": specifies the Infoblox network view to use for this network. This is a Infoblox IPAM driver specific attribute.
Other Infoblox specific attributes that are not shown in the example configuration:
- "network-container"
- "prefix-length": Instead of specifying a "subnet", the driver can be instructed to allocate a network of prefix length (integer) from within a network container (CIDR).
- "socket-dir": specifies an alternate directory where the socket file to send IPAM Daemon request to is located.
The default is ```/run/cni```.

Infoblox IPAM Driver Configuration
----------------------------------
The Infoblox IPAM Driver is comprised of two components:
- Infoblox IPAM Plugin (infoblox):
  This is the plugin executable specified as the IPAM type in the netconf. This is executed by CNI as a network

plugin and, by default in a rkt environment, is located in the ```/usr/lib/rkt/plugins/net``` directory.
- Infoblox IPAM Daemon (infoblox-cni-daemon):
  This is the component that interfaces with Infoblox to perform the IPAM functions. This is typically deployed
as a container and run as a service.

Running the IPAM Daemon
-----------------------
The IPAM Daemon accepts the following command line arguments, which specifies Infoblox Grid settings, IPAM Driver
settings and IPAM Policy settings respectively. Each one of the IPAM Policy settings is the fallback that take
effect when the same setting have not been specified in the network configuration file.

```
## Infoblox Grid Settings ##
--grid-host string
	IP of Infoblox Grid Host (default "192.168.124.200")
--wapi-port string
	Infoblox WAPI Port (default "443")
--wapi-username string
	Infoblox WAPI Username (default "")
--wapi-password string
	Infoblox WAPI Password (default "")
--wapi-version string
	Infoblox WAPI Version (default "2.3")

--ssl-verify string
	Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate. (default "false")

## IPAM Driver Settings ##
--socket-dir string
	Directory in which Infobox IPAM daemon socket is created (default "/run/cni")
--driver-name string
	Name of the IPAM driver. This is the file name used to create Infoblox IPAM daemon socket, and has to match the name specified as IPAM type in the CNI configuration. (default "infoblox")

## IPAM Policy Settings ##
--network-view string
	Infoblox Network View (default "default")
--network-container string
	Subnets will be allocated from this container if subnet is not specified in network config file (default "172.18.0.0/16")
--prefix-length integer
	The CIDR prefix length when allocating a subnet from Network Container (default 24)
```
NOTE:WAPI Version should be 2.3 or above

It is recommended that the Infoblox IPAM Daemon be run as a container. A docker image is availabe in Docker Hub
(infoblox/infoblox-cni-daemon). A skeleton shell script (run-rkt-daemon.sh) to run the docker image using rkt is
included. The shell script need to be executed with root permission.

Various ways to run the daemon include:
- run-rkt-daemon.sh:
  Runs the infoblox-cni-daemon docker image under rkt
- run-docker-daemon.sh:
  Runs the infoblox-cni-daemon docker image as a docker container.
- run-aci-daemon.sh:
  Runs a infoblox-cni-daemon ACI image under rkt.
- run-daemon.sh:
  Runs the infoblox-cni-daemon as a native exectuable.

Usage
-----
For a detailed description of an example use of the Infoblox IPAM Daemon in multi host rkt deployment, refer to
https://community.infoblox.com/t5/Community-Blog/CNI-Networking-and-IPAM/ba-p/7828

Before you can start using the driver, the Infobblox IPAM Daemon must be started using one of the methods
described in the section "Running the IPAM Daemon" above.

Assuming that you have deployed the example network configuration file (10-net-1.conf) shown in the
"CNI Configuration", which specifies the configuration for a network called "net-1", the following command starts a
rkt container attaching to the "net-1" network:

```
rkt run --interactive --net=net-1 quay.io/fermayo/ubuntu
```

When the container comes up, verify using the "ifconfig" command that IP has been successfully provisioned
from Infoblox.

Note
-----
This plugin supports CNI version 0.5.2 https://github.com/containernetworking/cni/tree/v0.5.2

Known bug
---------
With the Rocket(rkt) deallocation of IP does not work. This is due to unavailability of a feature in Rocket. https://github.com/infobloxopen/cni-infoblox/pull/10
