# Requires DOCKERHUB_ID and VERSION environment variables
ifeq ($(DOCKERHUB_ID),)
    ifeq ($(VERSION),)
        DAEMON_IMAGE=infoblox-cni-daemon
        PLUGIN_IMAGE=infoblox-cni-install
    else
        DAEMON_IMAGE=infoblox-cni-daemon:${VERSION}
        PLUGIN_IMAGE=infoblox-cni-install:${VERSION}
    endif
else
    ifeq ($(VERSION),)
        DAEMON_IMAGE=${DOCKERHUB_ID}/infoblox-cni-daemon
        PLUGIN_IMAGE=${DOCKERHUB_ID}/infoblox-cni-install
    else
        DAEMON_IMAGE=${DOCKERHUB_ID}/infoblox-cni-daemon:${VERSION}
        PLUGIN_IMAGE=${DOCKERHUB_ID}/infoblox-cni-install:${VERSION}
    endif
endif

# Delete local docker images
clean:
	docker rmi -f ${DAEMON_IMAGE} || true
	docker rmi -f ${PLUGIN_IMAGE} || true
			
# Ensure go dependencies
deps:
	dep ensure

# Build container Images...

build: clean deps
	docker build -t $(DAEMON_IMAGE) -f Dockerfile-cni-daemon .
	docker build -t $(PLUGIN_IMAGE) -f Dockerfile-cni-plugin-installer .

# Push image to user's docker hub. NOTE: requires DOCKERHUB_ID environment variable
push: build
	docker push $(DAEMON_IMAGE)
	docker push $(PLUGIN_IMAGE)
