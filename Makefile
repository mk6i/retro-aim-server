GO_RELEASER_CROSS_VERSION := v1.22.1
DOCKER_IMAGE_TAG := goreleaser-cross-garble:${GO_RELEASER_CROSS_VERSION}
GARBLE_VERSION := v0.12.1

DOCKER_RUN := @docker run \
	--env CGO_ENABLED=1 \
	--env GITHUB_TOKEN=$(GITHUB_TOKEN) \
	--rm \
	--volume `pwd`:/go/src/retro-aim-server \
	--workdir /go/src/retro-aim-server \
	$(DOCKER_IMAGE_TAG)

.PHONY: release
release:
	$(DOCKER_RUN) --clean

.PHONY: release-dry-run
release-dry-run:
	$(DOCKER_RUN) --clean --skip=validate --skip=publish

.PHONY: goreleaser-docker
goreleaser-docker:
	@docker build \
		--build-arg GARBLE_VERSION=$(GARBLE_VERSION) \
		--build-arg GO_RELEASER_CROSS_VERSION=$(GO_RELEASER_CROSS_VERSION) \
		--file Dockerfile.goreleaser \
		--tag $(DOCKER_IMAGE_TAG) \
		.