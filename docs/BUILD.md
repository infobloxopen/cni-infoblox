Building the Infoblox IPAM Driver
=================================

Ensure Dependancy
-----------------
To ensure dependancy use the following command
```
make deps
```

Building Image
--------------
To build the images use the following command:
```
make build
```

Pushing Image to Docker Hub
-------------------------------------
The Makefile also includes a build target to push the image to your Docker Hub.
To do that, you need to first setup the following environment variable:
```
export DOCKERHUB_ID="your-docker-hub-id"
export VERSION="your-version"

```
You can then use the following command to push the image to your Docker Hub:

To Push image
```
make push
```
Delete local docker images
-------------------------------------

make clean
```