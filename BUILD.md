Building the Infoblox IPAM Driver
=================================

Prerequisite
------------
1. golang development environment is installed (https://golang.org/doc/install)


Install Dependency
------------------
The driver primarily depends on ```github.com/containernetworking/cni``` and
```github.com/infobloxopen/infoblox-go-client```. They can be installed using the following commands:

```
go get github.com/containernetworking/cni
go get github.com/infobloxopen/infoblox-go-client
```

```infoblox-go-client``` is used by the IPAM Daemon to interact with Infoblox.

Build Executable
----------------
A Makefile is provided for automating the build process. The default target ```all``` builds the following binaries:

- infoblox-plugin:
  This is the plugin executable. This is typlically deployed in ```/usr/lib/rkt/plugins/net```, and has to be renamed
to match the plugin type, typically ``infoblox``, specified in network configuration.
- infoblox-cni-daemon:
  This is the IPAM daemon executable.

The Makefile also includes the following targets:

- docker-image:
  Builds docker images ```infoblox-cni-daemon``` and ```infoblox-cni-install```
- aci-image:
  Builds ACI image ```infoblox-cni-daemon.aci```
- images:
  Builds both docker-images and aci-image


Push Container Image to Docker Hub
----------------------------------
The Makefile includes a build target to push the ```infoblox-cni-daemon``` and ```infoblox-cni-install``` container images to your Docker Hub.
To do that, you need to first setup the following environment variable:

```
export DOCKERHUB_ID="your-docker-hub-id"

```
You can then use the following command to push the "ipam-driver" image to your Docker Hub:

```
make push
```
