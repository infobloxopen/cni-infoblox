CNI IPAM Driver for Kubernetes
==============================

Infoblox IPAM Driver for CNI
----------------------------

cni-infoblox is an IPAM driver for CNI that interfaces with Infoblox to provide IP Address Management
service. CNI is the generic plugin-based networking layer for supporting container runtime environments,
of which kubernetes is one.

For a detailed description of the driver, including a step by step deployment example, refer to the
"CNI Networking and IPAM" community blog on the Infolox website:
https://community.infoblox.com/t5/Community-Blog/CNI-Networking-and-IPAM/ba-p/7828

Prerequisite
------------
To use the plugin, you need access to the Infoblox DDI product. For evaluation purposes, you can download a
virtual version of the product from the Infoblox Download Center (https://www.infoblox.com/infoblox-download-center)
Alternatively, if you are an existing Infoblox customer, you can download it from the support site.

Refer to CONFIG.md for details on vNIOS configuration.

For setting up a kubernetes cluster one can use kubeadm which is designed to be a simple way for new users to start 
trying Kubernetes out. The following links can be useful.
Installing Kubeadm - https://kubernetes.io/docs/setup/independent/install-kubeadm/
Creating cluster - https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/ .

Versions Used
-------------
The versions installed in each node of kubernetes cluster for testing is:
```
Host - Ubuntu 16.04.3 LTS (GNU/Linux 4.4.0-87-generic x86_64)
docker - 1.13.1-0ubuntu1~16.04.2
kubeadm - 1.9.0-00
kubectl - 1.9.0-00
kubelet - 1.9.0-00
kubernetes-cni - 0.6.0-00

CNI source used to build the plugin and daemon - 0.6.0
Wapi version - 2.3
```

Build
-----
For dependencies and build instructions, refer to ```BUILD.md```.


CNI Configuration
-----------------
This section concerns only with CNI network configuration as it relates to the Infoblox IPAM Driver.
For details on CNI configuration in general, refer to https://github.com/containernetworking/cni/blob/master/README.md

To instruct CNI to execute the Infoblox IPAM plugin for a particular network, specify "infoblox" as the IPAM "type"
in the CNI network configuration file (netconf). CNI configuration files in a kubernetes environment is typically
located in ```/etc/cni/net.d``` . If there are multiple CNI configuration files in the directory, the first one in 
lexicographic order of file name is used. So make sure to name the netconf file with proper order. 

For example (/etc/cni/net.d/01-infoblox-ipam.conf):

```
{
    "name": "infoblox-ipam",
    "type": "macvlan",
    "master": "eth0",
    "ipam": {
        "type": "infoblox",
        "subnet": "10.0.0.0/24",

                "routes": [
                        {"dst": "172.18.0.0/24",
                         "gw":"10.0.0.1"}
                ],
                "network-view": "cni_view"
    }
}
```
Note : The following type of networks are supported out of the box:
 ```
       host-local
       bridge
       ipvlan
       macvlan
```  
To run macvlan, the promiscuous mode in eth0 should be enabled on each node. The following command
can be used to achieve it : ``ip link set eth0 promisc on``

Infoblox IPAM Driver Configuration
----------------------------------
The Infoblox IPAM Driver is comprised of two components:
- Infoblox IPAM Plugin (infoblox):
  This is the plugin executable specified as the IPAM type in the netconf. This is executed by CNI as a network
plugin and, by default in a kubernetes environment, is located in the ```/opt/cni/bin``` directory.
- Infoblox IPAM Daemon (infoblox-cni-daemon):
  This is the component that interfaces with Infoblox to perform the IPAM functions. This is typically deployed
as a kubernetes daemonset on each node.

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
	Infoblox WAPI Version (default "2.0")
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
	Subnets will be allocated from this container if subnet is not specified in network config file (default "172.18.0.0/16") . To have multiple subnet add comma separated subnet. (ex. "192.168.0.0/24,192.169.0.0/24")
--prefix-length integer
	The CIDR prefix length when allocating a subnet from Network Container (default 24)
```

It is recommended that the Infoblox IPAM Daemon be run as a daemonset in kubernetes cluster. A yaml file (infoblox-daemonset.yaml) is used to create the daemonset in kubernetes cluster
and can be done by the following command : ``kubectl create -f infoblox-daemonset.yaml`` . The daemonset should be created
before starting the driver. A docker image is available in Docker Hub, which packages the daemon binary in an image (infoblox/infoblox-cni-daemon) and used by the yaml file.

Usage
-----
For a detailed description of an example, which is more of an Infoblox IPAM Daemon in multi host rkt deployment(not in kubernetes), refer to
https://community.infoblox.com/t5/Community-Blog/CNI-Networking-and-IPAM/ba-p/7828

To use the driver start the daemonset as described in the section "Running the IPAM Daemon" above. Put the netconf file and plugin binary
in specified location as described in "CNI Configuration" and "Infoblox IPAM Driver Configuration" section respectively.

Test the pod connectivity by deploying apps in the kubernetes cluster.
For example : 
```
    #vi test-app.yaml
    apiVersion: apps/v1beta1
    kind: Deployment
    metadata:
      name: test-infoblox-deployment
    spec:
      replicas: 2
      template:
        metadata:
          labels:
            app: test-infoblox
        spec:
          containers:
          - name: test-infoblox
            image: ianneub/network-tools
            command: ["/bin/sh"]
            args: ["-c", "sleep 10000; echo 'I m dying' "]

```
```
kubectl create -f test-app.yaml
```
The command above starts test-infoblox-deployment with two pods. 

When the pods comes up, verify using the "ifconfig" inside the pod to check that IP has been successfully provisioned
from Infoblox. To verify the pod connectivity, ping the 2nd pod from inside the 1st pod. 