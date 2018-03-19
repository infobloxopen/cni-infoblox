
DAEMON_DOCKER_IMAGE=infoblox-cni-daemon
INSTALL_DOCKER_IMAGE=infoblox-cni-install
DEV_IMAGE=$(DOCKERHUB_ID)/$(DAEMON_DOCKER_IMAGE):$(VERSION)  # Requires DOCKERHUB_ID and VERSION environment variable
INSTALL_DEV_IMAGE=$(DOCKERHUB_ID)/$(INSTALL_DOCKER_IMAGE):$(VERSION)  # Requires DOCKERHUB_ID and VERSION environment variable
RELEASE_IMAGE=infoblox/$(DAEMON_DOCKER_IMAGE):$(VERSION) # Requires VERSION environment variable
INSTALL_RELEASE_IMAGE=infoblox/$(INSTALL_DOCKER_IMAGE):$(VERSION) # Requires VERSION environment variable

# Container Images...

docker-image:
	docker build -t $(DAEMON_DOCKER_IMAGE) -f Dockerfile-cni-daemon .
	docker build -t $(INSTALL_DOCKER_IMAGE) -f Dockerfile-cni-plugin-installer .

# Push image to user's docker hub. NOTE: requires DOCKERHUB_ID environment variable
push: docker-image
	docker tag $(DAEMON_DOCKER_IMAGE) $(DEV_IMAGE)
	docker push $(DEV_IMAGE)
	docker tag $(INSTALL_DOCKER_IMAGE) $(INSTALL_DEV_IMAGE)
	docker push $(INSTALL_DEV_IMAGE)

# Push image to infoblox docker hub
push-release: docker-image
	docker tag $(DAEMON_DOCKER_IMAGE) $(RELEASE_IMAGE)
	docker push $(RELEASE_IMAGE)
	docker tag $(INSTALL_DOCKER_IMAGE) $(INSTALL_RELEASE_IMAGE)
	docker push $(INSTALL_RELEASE_IMAGE)

# Clean everything
clean-all: clean clean-images

# Delete local docker images
clean-images:
	docker rmi -f $(DAEMON_DOCKER_IMAGE)
	/bin/rm -f $(DAEMON_ACI_IMAGE)

# Ensure go dependencies
deps:
	dep ensure