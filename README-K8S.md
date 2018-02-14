CNI IPAM Driver for Kubernetes
==============================

Cluster setup
-------------

For setting up a kubernetes cluster one can use kubeadm which is designed to be a simple way for new users to start 
trying Kubernetes out. The following links can be useful.
[Install Kubeadm](https://kubernetes.io/docs/setup/independent/install-kubeadm) and
[Create cluster](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm).

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

CNI Configuration
-----------------
This section concerns only with CNI network configuration as it relates to the Infoblox IPAM Driver.
For details on CNI configuration in general, refer [here](https://github.com/containernetworking/cni/blob/master/README.md).

To instruct CNI to execute the Infoblox IPAM plugin for a particular network, specify "infoblox" as the IPAM "type"
in the CNI network configuration file (netconf). CNI configuration files in a kubernetes environment is typically
located in ```/etc/cni/net.d``` . 

For example (/etc/cni/net.d/00-infoblox-ipam.conf):

```
{
    "name": "infoblox-ipam",
    "type": "macvlan",
    "master": "eth0",
    "ipam": {
        "type": "infoblox",
        "subnet": "10.0.0.0/24",
        "gateway": "10.0.0.1",
        "routes": [
                {"dst": "0.0.0.0"}
                ],
        "network-view": "cni_view"
        "prefix-length":"24"
    }
}
```
Note : To run macvlan network, the promiscuous mode in master interface(say eth0) should be enabled on each node in kubernetes cluster.
It can be done by the command ``ip link set eth0 promisc on`` . The promiscuous mode should be enabled on the network which is used by the kubernetes cluster nodes.

The following are the IPAM attributes:
- "type" (Required): specifies the plugin type and is also the file name of the plugin executable.
- "subnet" (Optional): specifies the CIDR to be used for the network. This is a well-known CNI attribute and is used by the driver.
- "gateway" (Optional): specifies the gateway for the network. This is a well-known CNI attribute and is simply passed through to CNI.

Note: if subnet is not provided in the conf.then, user needs to follow gateway format as given below:
a) if default prefix-length(24) used then gateway will be in 0.0.0.x format.
b) if prefix-length is provided then user should pass gateway in a away that new created subnet using network-container(default/user configured) should contain the gateway IP.
c) for example if prefix length used as 18, then user can give gateway in 0.0.y.x format. but gateway should lies in 255.255.192.0 Netmask.

- "routes" (Optional): specifies the routes for the network. This is a well-known CNI attribute and is simply passed through to CNI.
- "network-view" (Optional): specifies the Infoblox network view to use for this network. This is a Infoblox IPAM driver specific attribute.
Other Infoblox specific attributes that are not shown in the example configuration:
- "network-container" (Optional):Subnets will be allocated from this container if subnet is not specified in network config file(default "172.18.0.0/16").To have multiple subnet add comma separated subnet. (ex. "192.168.0.0/24,192.169.0.0/24")
- "prefix-length" (Optional): Instead of specifying a "subnet", the driver can be instructed to allocate a network of prefix length (integer) from within a network container (CIDR). 

Infoblox CNI IPAM Plugin 
========================

Features
--------
- Implementation of config map to enable automatic deployment of network configuration file and plugin on each node.

- User can give gateway in the format of 0.0.0.x when subnet not giving through the configuration file.

- Given type of networks are supported out of the box: ``bridge, ipvlan, macvlan`` . 
  
Limitations
-------
- Currently only supports IPv4 notIPv6.
- Network configuration file name should not be change(00infoblox-ipam.conf).  

Plugin Components
---------------

- Infoblox IPAM Plugin (infoblox):
  This is the plugin executable specified as the IPAM type in the netconf. This is executed by CNI as a network
plugin and, by default in a kubernetes environment, is located in the ```/opt/cni/bin``` directory.
- Infoblox IPAM Daemon (infoblox-cni-daemon):
  This is the component that interfaces with Infoblox to perform the IPAM functions. This is typically deployed
as a kubernetes daemonset on each node.

IPAM Daemon Configuration
------------------------
The IPAM Daemon accepts the following command line arguments, which specifies Infoblox Grid settings, IPAM Driver
settings and IPAM Policy settings respectively. Each one of the IPAM Policy settings is the fallback that take
effect when the same setting have not been specified in the network configuration file. The following settings
can be configured in the file ``infoblox-daemonset.yaml`` .

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
NOTE:WAPI Version should be 2.3 or above


How do we install Infoblox CNI Plugin ?
--------------------------------------
```
    kubectl create -f infoblox-daemonset.yaml
```
It is recommended that the Infoblox IPAM Daemon be run as a daemonset in kubernetes cluster.
The daemonset should be created before starting the plugin. A docker image is available in Docker Hub, which packages the daemon binary in an image (infoblox/infoblox-cni-daemon) and used by the infoblox-daemonset.yaml file.
 
 ```
     kubectl create -f infoblox-cni-install.yaml
 ``` 
The above command will create a daemonset in kubernetes cluster. It will install infoblox plugin binary and network configuration file
in the locations ``/opt/cni/bin`` and ``/etc/cni/net.d`` respectively. A docker image is available in Docker Hub, which packages the daemon binary in an image (infoblox/infoblox-cni-install) and used by the yaml file.
For making any changes in network configuration we can change the network config file contents part in the  infoblox-cni-install.yaml(shown below) and do 
``kubectl apply -f infoblox-cni-install.yaml`` or by changing the configmap ``kubectl edit configmap infoblox-cni-cfg --namespace=kube-system``.

NOTE: It takes approx. 1 minute to reflect the configmap changes using configmap edit command.

```
ipam_conf_file_name: infoblox-ipam.conf
  ## Network Config file contents##
  ##This key should match the value of the key 'ipam_conf_file_name'##
  infoblox-ipam.conf: |
    {
    "name": "ipam-test",
    "type": "macvlan",
    "master":"eth0",
    "ipam": {
        "type": "infoblox",
        "subnet": "10.0.0.0/24",
        "gateway":"10.0.0.1",
        "network-view": "cni_view"
        }
    }
```

Note:- If there are multiple CNI configuration files in the kubernetes network config directory(i.e. /etc/cni/net.d), then the first one in 
       lexicographic order of file name is used. So make sure to name the network configuration file with proper order. In the above example
       filename is given as  ```infoblox-ipam.conf``` which should match the value of the key ```ipam_conf_file_name```.

Usage
-----
For a detailed description of an example, which is more of an Infoblox IPAM Daemon in multi host rkt deployment(not in kubernetes), refer [here](https://community.infoblox.com/t5/Community-Blog/CNI-Networking-and-IPAM/ba-p/7828).

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

When the pods comes up, verify using the "ifconfig" inside the pod to check that IP has been successfully provisioned from Infoblox. 
To verify the pod connectivity, ping the 2nd pod from inside the 1st pod.
