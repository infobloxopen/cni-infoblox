# Requires DOCKERHUB_ID and VERSION environment variables
ifeq ($(DOCKERHUB_ID),)
    ifeq ($(VERSION),)
        DAEMON_IMAGE=cni-infoblox-daemon
        PLUGIN_IMAGE=cni-infoblox-plugin
    else
        DAEMON_IMAGE=cni-infoblox-daemon:${VERSION}
        PLUGIN_IMAGE=cni-infoblox-plugin:${VERSION}
    endif
else
    ifeq ($(VERSION),)
        DAEMON_IMAGE=${DOCKERHUB_ID}/cni-infoblox-daemon
        PLUGIN_IMAGE=${DOCKERHUB_ID}/cni-infoblox-plugin
    else
        DAEMON_IMAGE=${DOCKERHUB_ID}/cni-infoblox-daemon:${VERSION}
        PLUGIN_IMAGE=${DOCKERHUB_ID}/cni-infoblox-plugin:${VERSION}
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
	docker build -t $(DAEMON_IMAGE) -f Dockerfile-infoblox-daemon .
	docker build -t $(PLUGIN_IMAGE) -f Dockerfile-infoblox-plugin .

# Push image to user's docker hub. NOTE: requires DOCKERHUB_ID environment variable
push: build
	docker push $(DAEMON_IMAGE)
	docker push $(PLUGIN_IMAGE)
