COMMON_SOURCES=config.go driver-socket.go infoblox-skel.go
PLUGIN_SOURCES=infoblox-plugin.go $(COMMON_SOURCES)
DAEMON_SOURCES=infoblox-daemon.go infoblox-ipam.go $(COMMON_SOURCES)
PLUGIN_BINARY=infoblox-plugin
DAEMON_BINARY=infoblox-daemon
DAEMON_IMAGE=infoblox-daemon

LOCAL_IMAGE=$(DAEMON_IMAGE)

# Build binary - this is the default target
build: $(PLUGIN_BINARY) $(DAEMON_BINARY)


# Build binary and docker image
all: build


# Build local docker image
image: build
	docker build -t $(LOCAL_IMAGE) .

# Push image to user's docker hub. NOTE: requires DOCKERHUB_ID environment variable
push: image
	docker tag $(LOCAL_IMAGE) $(DEV_IMAGE)
	docker push $(DEV_IMAGE)

# Push image to infoblox docker hub
push-release: image
	docker tag $(LOCAL_IMAGE) $(RELEASE_IMAGE)
	docker push $(RELEASE_IMAGE)

$(PLUGIN_BINARY): $(PLUGIN_SOURCES)
	go build -o $(PLUGIN_BINARY) $(PLUGIN_SOURCES)

$(DAEMON_BINARY): $(DAEMON_SOURCES)
	go build -o $(DAEMON_BINARY) $(DAEMON_SOURCES)

# Delete binary for clean build
clean:
	rm -f $(PLUGIN_BINARY) $(DAEMON_BINARY)

# Delete local docker images
clean-images:
	docker rmi -f $(LOCAL_IMAGE) $(DEV_IMAGE) $(RELEASE_IMAGE)

# Clean everything
clean-all: clean clean-images
