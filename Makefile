################################################################################
# Build & release helpers
################################################################################

DOCKER_IMAGE_TAG_GO_RELEASER := goreleaser/goreleaser:v2.9.0
DOCKER_RUN_GO_RELEASER := @docker run \
	--env CGO_ENABLED=0 \
	--env GITHUB_TOKEN=$(GITHUB_TOKEN) \
	--rm \
	--volume `pwd`:/go/src/retro-aim-server \
	--workdir /go/src/retro-aim-server \
	$(DOCKER_IMAGE_TAG_GO_RELEASER)
OSCAR_HOST ?= ras.dev

.PHONY: config
config: ## Generate config file template from Config struct
	go generate ./config

.PHONY: release
release: ## Run a clean, full GoReleaser run (publish + validate)
	$(DOCKER_RUN_GO_RELEASER) --clean

.PHONY: release-dry-run
release-dry-run: ## GoReleaser dry-run (skips validate & publish)
	$(DOCKER_RUN_GO_RELEASER) --clean --skip=validate --skip=publish

.PHONY: docker-image-ras
docker-image-ras: ## Build Retro AIM Server image
	docker build -t ras:latest -f Dockerfile .

.PHONY: docker-image-stunnel
docker-image-stunnel: ## Build stunnel image pinned to v5.75 / OpenSSL 1.0.2u
	docker build -t ras-stunnel:5.75-openssl-1.0.2u -f Dockerfile.stunnel .

.PHONY: docker-image-certgen
docker-image-certgen: ## Build minimal helper image with openssl & nss tools
	docker build -t ras-certgen:latest -f Dockerfile.certgen .

.PHONY: docker-images
docker-images: docker-image-ras docker-image-stunnel docker-image-certgen

.PHONY: docker-run
docker-run:
	OSCAR_HOST=$(OSCAR_HOST) docker compose up

################################################################################
# SSL Helpers
################################################################################

.PHONY: docker-certs
docker-certs: clean-certs ## Create SSL certificates for AIM 6.0+ clients
	OSCAR_HOST=$(OSCAR_HOST) docker compose run --rm cert-gen

.PHONY: clean-certs
clean-certs: ## Remove all generated certificates & NSS DB
	rm -rf certs/*
