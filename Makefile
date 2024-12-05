DOCKER_IMAGE_TAG := goreleaser/goreleaser:v2.4.8

DOCKER_RUN := @docker run \
	--env CGO_ENABLED=0 \
	--env GITHUB_TOKEN=$(GITHUB_TOKEN) \
	--rm \
	--volume `pwd`:/go/src/retro-aim-server \
	--workdir /go/src/retro-aim-server \
	$(DOCKER_IMAGE_TAG)

.PHONY: config
config:
	go generate ./config

.PHONY: release
release:
	$(DOCKER_RUN) --clean

.PHONY: release-dry-run
release-dry-run:
	$(DOCKER_RUN) --clean --skip=validate --skip=publish