CNI IPAM Driver
===============

Infoblox IPAM Driver for CNI
----------------------------

cni-infoblox is an IPAM driver for CNI that interfaces with Infoblox to provide IP Address Management
service. CNI is the generic plugin-based networking layer for supporting container runtime environments.

For a detailed description of the driver, including a step by step deployment example, refer to the community blog on the Infoblox website: [CNI Networking and IPAM](https://community.infoblox.com/t5/Community-Blog/CNI-Networking-and-IPAM/ba-p/7828).

Prerequisite
------------

A NIOS DDI Appliance with cloud automation License.

To use the plugin, you need access to the Infoblox DDI product. For evaluation purposes, you can download a
virtual version of the product from the [Infoblox Download Center](https://www.infoblox.com/infoblox-download-center).
Alternatively, if you are an existing Infoblox customer, you can download it from the support site and you can also assign temp license by login into the Infoblox DDI appliance console with this command ```set temp_license```.

Refer to [CONFIG.md](CONFIG.md) for details on vNIOS configuration.

Configuring Supported container runtimes
----------------------------------------

Refer to the following links to configure each container runtime to use infoblox cni plugin:

* Kubernetes - [README-K8S.md](README-K8S.md)
* Rocket - With the Rocket(rkt), deallocate of IP does not work. Until rocket has latest cni, the infoblox plugin 
support will not be provided. Still configuring rocket to use infoblox cni plugin can be read at [README-rkt.md](README-rkt.md)

Development
-----------

* Build - For dependencies and build instructions, refer to [BUILD.md](BUILD.md) .