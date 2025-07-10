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

.PHONY: config
config: ## Generate config file template from Config struct
	go generate ./config

.PHONY: release
release: ## Run a clean, full GoReleaser run (publish + validate)
	$(DOCKER_RUN_GO_RELEASER) --clean

.PHONY: release-dry-run
release-dry-run: ## GoReleaser dry-run (skips validate & publish)
	$(DOCKER_RUN_GO_RELEASER) --clean --skip=validate --skip=publish

.PHONY: ras-image
ras-image: ## Build Retro AIM Server image
	docker build -t ras:main -f Dockerfile .

################################################################################
# SSL Helpers
################################################################################

CERT_DIR       ?= certs
CERT_GEN_IMAGE ?= cert-nss
CERT_NAME      ?= ras.dev
CERT_NSSDB_DIR := $(CERT_DIR)/nssdb
CERT_PEM       := $(CERT_DIR)/server.pem

.PHONY: stunnel-image
stunnel-image: ## Build stunnel image pinned to v5.75 / OpenSSL 1.0.2u
	docker build -t stunnel:5.75-openssl-1.0.2u -f Dockerfile.stunnel .

.PHONY: cert-gen-image
cert-gen-image: ## Build minimal helper image with openssl & nss tools
	docker build -t $(CERT_GEN_IMAGE) -f Dockerfile.certgen .

.PHONY: certs
certs: clean-certs cert-gen-image ## Create SSL certificates for AIM 6.0+ clients
	mkdir -p $(CERT_DIR)

 	# create SSL certificate
	docker run --rm -v "$$PWD":/work -w /work/$(CERT_DIR) $(CERT_GEN_IMAGE) \
		openssl req -x509 -newkey rsa:1024 \
			-keyout "key.pem" \
			-out "cert.pem" \
			-sha256 -days 365 -nodes \
			-subj "/CN=$(CERT_NAME)"
	cat $(CERT_DIR)/cert.pem $(CERT_DIR)/key.pem > $(CERT_PEM)
	rm $(CERT_DIR)/cert.pem $(CERT_DIR)/key.pem

 	# build NSS DB
	mkdir -p $(CERT_NSSDB_DIR)
	docker run -it --rm -v "$$PWD":/work -w /work $(CERT_GEN_IMAGE) \
		sh -c "certutil -N -d $(CERT_NSSDB_DIR) --empty-password && \
		       certutil -A -n 'RAS' -t 'CT,,C' -i $(CERT_PEM) -d $(CERT_NSSDB_DIR)"

	@echo "Successfully created certificates in '$(CERT_DIR)/'"

clean-certs: ## Remove all generated certificates & NSS DB
	rm -rf $(CERT_DIR)