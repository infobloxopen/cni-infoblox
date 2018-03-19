# Requires DOCKERHUB_ID and VERSION environment variables
DAEMON_NAME=infoblox-cni-daemon
PLUGIN_NAME=infoblox-cni-install
DAEMON_IMAGE=$(DOCKERHUB_ID)/$(DAEMON_NAME):$(VERSION) 
PLUGIN_IMAGE=$(DOCKERHUB_ID)/$(PLUGIN_NAME):$(VERSION)

# Container Images...

docker-image:
	docker build -t $(DAEMON_IMAGE) -f Dockerfile-cni-daemon .
	docker build -t $(PLUGIN_IMAGE) -f Dockerfile-cni-plugin-installer .

# Push image to user's docker hub. NOTE: requires DOCKERHUB_ID environment variable
push: docker-image
	docker push $(DAEMON_IMAGE)
	docker push $(PLUGIN_IMAGE)

# Delete local docker images
clean-images:
	docker rmi ${DAEMON_IMAGE} || true
	docker rmi ${PLUGIN_IMAGE} || true
			
# Ensure go dependencies
deps:
	dep ensure