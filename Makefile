GO_RELEASER_VERSION := v1.21.5
DOCKER_RUN := @docker run \
	--rm \
	-e CGO_ENABLED=1 \
	-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
	-v `pwd`:/go/src/retro-aim-server \
	-w /go/src/retro-aim-server \
	ghcr.io/goreleaser/goreleaser-cross:$(GO_RELEASER_VERSION)

.PHONY: release
release:
	$(DOCKER_RUN) --clean

.PHONY: release-dry-run
release-dry-run:
	$(DOCKER_RUN) --clean --skip=validate --skip=publish
