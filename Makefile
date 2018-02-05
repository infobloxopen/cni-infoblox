PLUGIN_DIR=plugin
DAEMON_DIR=daemon

COMMON_SOURCES=config.go driver-socket.go cmdargs.go infoblox-ipam.go
PLUGIN_SOURCES=$(PLUGIN_DIR)/plugin.go $(COMMON_SOURCES)
DAEMON_SOURCES=$(DAEMON_DIR)/daemon.go $(COMMON_SOURCES)

PLUGIN_BINARY=infoblox-plugin
DAEMON_BINARY=infoblox-cni-daemon
ALL_BINARIES=$(PLUGIN_BINARY) $(DAEMON_BINARY)

DAEMON_ACI_IMAGE=infoblox-cni-daemon.aci
DAEMON_DOCKER_IMAGE=infoblox-cni-daemon
INSTALL_DOCKER_IMAGE=infoblox-cni-install
DEV_IMAGE=$(DOCKERHUB_ID)/$(DAEMON_DOCKER_IMAGE)  # Requires DOCKERHUB_ID environment variable
INSTALL_DEV_IMAGE=$(DOCKERHUB_ID)/$(INSTALL_DOCKER_IMAGE)  # Requires DOCKERHUB_ID environment variable
RELEASE_IMAGE=infoblox/$(DAEMON_DOCKER_IMAGE)
INSTALL_RELEASE_IMAGE=infoblox/$(INSTALL_DOCKER_IMAGE)


# Build binary
all: build

# Build binary - this is the default target
build: $(ALL_BINARIES)

$(PLUGIN_BINARY): $(PLUGIN_SOURCES)
	cd $(PLUGIN_DIR); go build -o ../$(PLUGIN_BINARY)

$(DAEMON_BINARY):$(DAEMON_SOURCES)
	cd $(DAEMON_DIR); go build -o ../$(DAEMON_BINARY)

# Delete binary for clean build
clean:
	rm -f $(ALL_BINARIES)


# Container Images...

images: aci-image docker-image

docker-image: $(DAEMON_BINARY)
	docker build -t $(DAEMON_DOCKER_IMAGE) .
	docker build -t $(INSTALL_DOCKER_IMAGE) -f Dockerfile-install-cni .

# Push image to user's docker hub. NOTE: requires DOCKERHUB_ID environment variable
push: docker-image
	docker tag $(DAEMON_DOCKER_IMAGE) $(DEV_IMAGE)
	docker push $(DEV_IMAGE)

# Push image to infoblox docker hub
push-release: docker-image
	docker tag $(DAEMON_DOCKER_IMAGE) $(RELEASE_IMAGE)
	docker push $(RELEASE_IMAGE)

aci-image: $(DAEMON_ACI_IMAGE)

$(DAEMON_ACI_IMAGE): $(DAEMON_BINARY)
	./build-aci.sh

# Clean everything
clean-all: clean clean-images

# Delete local docker images
clean-images:
	docker rmi -f $(DAEMON_DOCKER_IMAGE)
	/bin/rm -f $(DAEMON_ACI_IMAGE)
