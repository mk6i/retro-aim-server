#!/usr/bin/env bash

# setup.sh
# Installer for Retro AIM Server Docker Setup

set -e

#
# boiler plate
#
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()    { echo -e "${CYAN}[*] $1${NC}"; }
success(){ echo -e "${GREEN}[✔] $1${NC}"; }
warn()   { echo -e "${YELLOW}[!] $1${NC}"; }
error()  { echo -e "${RED}[✖] $1${NC}"; exit 1; }

check_prereqs() {
    log "Checking for required tools..."
    for cmd in git docker make; do
        if ! command -v $cmd &> /dev/null; then
            error "Missing required command: $cmd"
        fi
    done
    success "All prerequisites are installed."
}

clone_repo() {
    if [ -d "retro-aim-server" ]; then
        warn "Directory 'retro-aim-server' already exists. Skipping clone."
    else
        log "Cloning Retro AIM Server repository..."
        git clone https://github.com/mk6i/retro-aim-server.git || error "Failed to clone repository"
    fi
    cd retro-aim-server || error "Failed to enter repo directory"
    success "Repository ready."
}

build_images() {
    log "Building Docker images..."
    make docker-images || error "Docker image build failed"
    success "Docker images built successfully."
}

setup_ssl_cert() {
    read -rp "$(echo -e ${YELLOW}"Enter the OSCAR_HOST (e.g., ras.dev): "${NC})" OSCAR_HOST < /dev/tty
    if [ -z "$OSCAR_HOST" ]; then
        error "OSCAR_HOST is required"
    fi

    log "SSL certificate options:"
    echo -e "${YELLOW}1) Generate self-signed certificate"
    echo -e "2) Use existing PEM certificate at certs/server.pem${NC}"
    read -rp "$(echo -e ${CYAN}"Choose an option [1/2]: "${NC})" cert_choice < /dev/tty

    if [ "$cert_choice" == "1" ]; then
        log "Generating self-signed SSL certificate for $OSCAR_HOST..."
        make docker-cert OSCAR_HOST="$OSCAR_HOST" || error "Failed to generate self-signed cert"
        success "Self-signed certificate generated."
    elif [ "$cert_choice" == "2" ]; then
        if [ ! -f "certs/server.pem" ]; then
            error "certs/server.pem not found. Please place your PEM file there and rerun."
        fi
        success "Using existing certificate at certs/server.pem."
    else
        error "Invalid option for certificate setup."
    fi
}

generate_nss() {
    log "Generating NSS certificate database..."
    make docker-nss || error "Failed to generate NSS cert DB"
    success "NSS certificate database created at certs/nss/"
}

start_server() {
    log "Starting Retro AIM Server with hostname $OSCAR_HOST..."
    make docker-run-bg OSCAR_HOST="$OSCAR_HOST" || error "Failed to start server"
    success "Retro AIM Server is running."
}

final_steps() {
    echo -e "${GREEN}
Retro AIM Server setup complete!
===================================
Next Steps:

1. Copy the NSS certs from 'certs/nss/' to each AIM 6.2+ client.

2. Ensure clients can resolve '${OSCAR_HOST}' to your server IP.
   Add the following to each client's hosts file if DNS isn't used:
   ${NC}
${YELLOW}127.0.0.1 $OSCAR_HOST${NC}
${GREEN}
3. For AIM 6.x client setup instructions, see:
   https://github.com/mk6i/retro-aim-server/blob/main/docs/AIM6.md#aim-6265312-setup

Enjoy!
${NC}"
}

main() {
    check_prereqs
    clone_repo
    build_images
    setup_ssl_cert
    generate_nss
    start_server
    final_steps
}

main "$@"

