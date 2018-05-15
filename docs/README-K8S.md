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
Wapi version - 2.5
```

CNI Configuration
-----------------
This section concerns only with CNI network configuration as it relates to the Infoblox IPAM Driver.
For details on CNI configuration in general, refer [here](https://github.com/containernetworking/cni/blob/master/README.md).

To instruct CNI to execute the Infoblox IPAM plugin for a particular network, specify "infoblox" as the IPAM "type"
in the CNI network configuration file (netconf). CNI configuration files in a kubernetes environment is typically
located in ```/etc/cni/net.d``` . 

For example (/etc/cni/net.d/infoblox-ipam.conf):

```
{
    "name": "infoblox-ipam-network",
    "type": "macvlan",
    "master": "eth0",
    "ipam": {
        "type": "infoblox",
        "subnet": "10.0.0.0/24",
        "gateway": "10.0.0.1",
        "network-view": "cni_view"
        
    }
}
```
Note : To run macvlan network, the promiscuous mode in master interface(say eth0) should be enabled on each nodes of the kubernetes cluster.
It can be done by the command ``ip link set eth0 promisc on`` . The promiscuous mode should be enabled on the network which is used by the kubernetes cluster nodes.

The following are the IPAM attributes:
- "type" (Required): specifies the plugin type and is also the file name of the plugin executable.
- "subnet" (Optional): specifies the CIDR to be used for the network. This is a well-known CNI attribute and is used by the driver.
- "gateway" (Optional): specifies the gateway for the network. This is a well-known CNI attribute and is simply passed through to CNI.

- "routes" (Optional): specifies the routes for the network. This is a well-known CNI attribute and is simply passed through to CNI.
- "network-view" (Optional): specifies the Infoblox network view to use for this network. This is a Infoblox IPAM driver specific attribute.
Other Infoblox specific attributes that are not shown in the example configuration:

Infoblox CNI IPAM Plugin 
========================

Features
--------
- Infoblox CNI plugin is a alternative for 'host-local' & 'dhcp' IPAM plugins.
- Used along with ``bridge, macvlan, ipvlan`` network types.
- Implementation of config map to enable automatic deployment of network configuration file and plugin on each node.
- User can give gateway in the format of 0.0.0.x when subnet not giving through the configuration file.

  
Limitations
-------
- Currently only supports IPv4 not IPv6.
- Kube-dns will not reach service ip in case of macvlan & ipvlan network types used.
- Need to create iptables rules and routes while using along with ``bridge`` network type.
- Network configuration file name should not be changed (00infoblox-ipam.conf).
- Open issues related to this.
	- [https://github.com/kubernetes/kubernetes/issues/53089]
	- [https://github.com/kubernetes/dns/issues/176]
	- [https://stackoverflow.com/questions/35695840/iptables-not-working-on-macvlan-traffic-in-container]


Plugin Components
---------------

**CNI Infoblox daemon:**
  This is the component that interfaces with Infoblox to perform the IPAM functions. This is typically deployed as a kubernetes daemonset (cni-infoblox-daemon) on each node.

**CNI Infoblox IPAM Plugin (infoblox):**
  This is the plugin executable specified as the IPAM type in the netconf. This is executed by CNI along with other network plugin it is located in the ```/opt/cni/bin``` directory. This is 
typically deployed as a kubernetes daemonset (cni-infoblox-plugin) on each node.

CNI Infoblox daemon Configuration
------------------------
This Infoblox daemon accepts the following command line arguments, which specifies Infoblox Grid settings, IPAM Driver settings and IPAM Policy settings respectively. Each one of these IPAM Policy settings is the fallback that take effect when the same setting have not been specified in the network configuration file. The following settings can be configured in the file ``cni-infoblox-daemon.yaml``.

```
## Infoblox Grid Settings ##
--grid-host string
	IP of Infoblox Grid Host (default "192.168.124.200")
--wapi-port string
	Infoblox WAPI Port (default "443")
--wapi-username string
	Infoblox WAPI Username (default "")
--wapi-version string
	Infoblox WAPI Version (default "2.5")
--ssl-verify string
	Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate. (default "false")
--cluster-name
    User defined cluster name to identify the deployment (default "cluster-1")

## IPAM Driver Settings ##
--socket-dir string
	Directory in which Infobox IPAM daemon socket is created (default "/run/cni")
--driver-name string
	Name of the IPAM driver. This is the file name used to create Infoblox IPAM daemon socket, and has to match the name specified as IPAM type in the CNI configuration. (default "infoblox")

## IPAM Policy Settings ##
--network-view string
	Infoblox Network View (default "default")
--network string
        Network cidr to be used to assign ip address for pods if cidr info is not provided in cni network conf file ( default "172.18.0.0/16" )
```

wapi-password should be passed via kubernetes secrets. Refer to [K8s-Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) for more details.

```
Assume 'infoblox' is your wapi password so base64 encode it.

$ echo -n "infoblox" | tr -d '\n' | base64
aW5mb2Jsb3g=

You have to update this base64 encoded value in k8s/cni-infoblox-daemon.yaml in below mentioned section.

---
apiVersion: v1
kind: Secret
metadata:
  name: infoblox-secret
  namespace: kube-system
type: Opaque
data:
  wapi-password: <UPDATE YOUR ENCODED VALUE>

```

NOTE: WAPI Version should be 2.5 or above


CNI Infoblox IPAM plugin Configuration
--------------------------------------

In this release Infoblox IPAM plugin supports following two types of IPAM configuration only possible to assign ip for pods


	1) **One CIDR across all nodes**
	2) **Unique CIDR on each node**


**One CIDR for pod across all nodes**

All CNI network conf must be identical across all nodes. This can be achieved by using this **k8s/cni-infoblox-plugin.yaml** file and update below section like below and deploy cni-infoblox-plugin daemonset. It will just copy the **infoblox** plugin binary and net conf mentioned in the config map section to all the worker nodes.

Assume my pod networks is **10.0.0.0/12**, then

```
  infoblox-ipam.conf: |
    {
    "name": "infoblox-ipam-network",
    "type": "macvlan",
    "master":"eth0",
    "ipam": {
        "type": "infoblox",
        "subnet": "10.0.0.0/12",
        "gateway":"10.0.0.1",
        "network-view": "flat_view"
        }
    }
```

or

A network automatically allocated by CNI Infoblox daemon if subnet is not metioned.

```
  infoblox-ipam.conf: |
    {
    "name": "infoblox-ipam-network",
    "type": "macvlan",
    "master":"eth0",
    "ipam": {
        "type": "infoblox",
        "network-view": "flat_view"
        }
    }
```

**Unique CIDR for pod on each node**

When opting for different pod cidrs in each worker node, in that case CNI network conf on each node will be different especially the below 3 parameters should be different 
- name
- subnet
- gateway

The name should be different because relevent network will be created in Infoblox applinaces so we can't have a network with same name and different subnet. The subnet/cidr and gatway are different because in background routes & iptables will be configured in each node for respective subnet/cidr only. 

User have to manually update the CNI network conf file. so use this **k8s/cni-infoblox-plugin-without-net-conf.yaml** file to deploy cni-infoblox-plugin daemonset only. It will just copy the **infoblox** plugin binary only to all the worker nodes.

Here are is an Example.

Node 1

```
  infoblox-ipam.conf: |
    {
    "name": "infoblox-ipam-network_10",
    "type": "macvlan",
    "master":"eth0",
    "ipam": {
        "type": "infoblox",
        "subnet": "192.168.10.0/24",
        "gateway":"192.168.10.1",
        "network-view": "node_view"
        }
    }
```

Node 2

```
  infoblox-ipam.conf: |
    {
    "name": "infoblox-ipam-network_11",
    "type": "macvlan",
    "master":"eth0",
    "ipam": {
        "type": "infoblox",
        "subnet": "192.168.11.0/24",
        "gateway":"192.168.11.1",
        "network-view": "node_view"
        }
    }
```




How do we install Infoblox CNI Plugin ?
--------------------------------------

**CNI Infoblox daemon**

```
    kubectl create -f k8s/cni-infoblox-daemon.yaml
```
cni-infoblox-daemon daemonset should be created before starting the plugin. It requied a docker image which is available at [infoblox/cni-infoblox-daemon](https://hub.docker.com/r/infoblox/cni-infoblox-daemon/) which has the binary in an image of cni-infoblox-daemon.

NOTE: Don't forget to update base64 encoded wapi-password in k8s/cni-infoblox-daemon.yaml

**CNI Infoblox plugin**
 
cni-infoblox-plugin daemonset can be deployed in 2 ways with,

	1) infoblox plugin + CNI network configuration file
		    kubectl create -f k8s/cni-infoblox-plugin.yaml
	2) infoblox plugin only
			kubectl create -f k8s/cni-infoblox-plugin-without-net-conf.yaml

Any of the above commands will create a cni-infoblox-plugin daemonset in kubernetes cluster. It required a docker image which is available at [infoblox/cni-infoblox-plugin](https://hub.docker.com/r/infoblox/cni-infoblox-plugin/) It will install infoblox plugin binary and network configuration file(if used k8s/cni-infoblox-plugin.yaml) in the locations ``/opt/cni/bin`` and ``/etc/cni/net.d`` respectively in all the worker nodes. If k8s/cni-infoblox-plugin-without-net-conf.yaml used it will copy infoblox plugin binary only. 

For making any changes in CNI network configuration we can change the network config file contents part in the cni-infoblox-plugin (shown below)

do this before creating the daemonset

``kubectl apply -f k8s/cni-infoblox-plugin.yaml``

or by changing the configmap once daemonset created

``kubectl edit configmap infoblox-cni-cfg --namespace=kube-system``

NOTE: It takes approx. 1 minute to reflect the configmap changes using configmap edit command.

```
ipam_conf_file_name: infoblox-ipam.conf
  ## Network Config file contents##
  ## This key should match the value of the key 'ipam_conf_file_name'##
  infoblox-ipam.conf: |
    {
    "name": "infoblox-ipam-network",
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

Note:- If there are multiple CNI configuration files in the kubernetes network config directory(i.e. /etc/cni/net.d), then the first one in lexicographic order of file name is used. So make sure to name the network configuration file with proper order. In the above example      filename is given as  ```infoblox-ipam.conf``` which should match the value of the key ```ipam_conf_file_name```.

Usage
-----
For a detailed description of an example, which is more of an Infoblox IPAM Daemon in multi host rkt deployment(not in kubernetes), refer [here](https://community.infoblox.com/t5/Community-Blog/CNI-Networking-and-IPAM/ba-p/7828).

To use the driver start the daemonset as described in the section "Running the IPAM Daemon" above. Put the netconf file and plugin binary
in specified location as described in "CNI Configuration" and "Infoblox IPAM Driver Configuration" section respectively.

Test the pod connectivity by deploying apps in the kubernetes cluster. use this below example and save it in *test-app.yaml*

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
kubectl create -f example/test-app.yaml
```
The command above starts test-infoblox-deployment with two pods. 

When the pods comes up, verify using the "ifconfig" inside the pod to check that IP has been successfully provisioned from Infoblox. 
To verify the pod connectivity, ping the 2nd pod from inside the 1st pod.

Use Existing Network
--------------------

By default required network view & network will be created based on configuration available at infoblox-ipam.conf. if you want to use existing network view & networks available at the NIOS update below Extensible Attributes (EA`s).
``Network Name, CMP Type, Cloud API Owned, Tenant ID ``.

For example: Assume your network name is "ABCNET"  in network configuration, Set EA`s like below.

```
"Network Name" =  "ABCNET"
"CMP Type" =  "Kubernetes"
"Cloud API Owned" =  "True"
"Tenant ID" =  "Testing"
```
