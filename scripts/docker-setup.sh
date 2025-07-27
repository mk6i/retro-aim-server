#!/usr/bin/env bash

# docker-setup.sh
# POSIX-compatible installer for Retro AIM Server Docker Setup

set -e

# ANSI color codes (fallback for POSIX shells that support it)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()    { printf "%b\n" "${CYAN}[*] $1${NC}"; }
success(){ printf "%b\n" "${GREEN}[✔] $1${NC}"; }
warn()   { printf "%b\n" "${YELLOW}[!] $1${NC}"; }
error()  { printf "%b\n" "${RED}[✖] $1${NC}"; return 1; }

check_os() {
    case "$(uname -s)" in
        Linux|Darwin)
            : # supported
            ;;
        *)
            error "Unsupported OS: $(uname -s)" || return 1
            ;;
    esac
}

check_prereqs() {
    log "Checking for required tools..."
    for cmd in git docker make; do
        command -v "$cmd" >/dev/null 2>&1 || error "Missing required command: $cmd" || return 1
    done
    success "All prerequisites are installed."
}

resolve_repo_root() {
    SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

    # If this script is inside a Git repo already, use that
    if git -C "$SCRIPT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        REPO_ROOT=$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)
    else
        TMP_DIR="retro-aim-server"
        if [ -d "$TMP_DIR/.git" ]; then
            warn "Directory '$TMP_DIR' is already a git repo. Using existing directory."
        else
            log "Cloning Retro AIM Server repository..."
            git clone https://github.com/mk6i/retro-aim-server.git "$TMP_DIR" || error "Failed to clone repository" || return 1
        fi
        REPO_ROOT=$(cd "$TMP_DIR" && pwd)
    fi
    cd "$REPO_ROOT" || error "Failed to enter repository directory" || return 1
}

build_images() {
    log "Building Docker images..."
    make docker-images || error "Docker image build failed" || return 1
    success "Docker images built successfully."
}

setup_ssl_cert() {
    printf "%b" "${YELLOW}Enter the OSCAR_HOST (e.g., ras.dev): ${NC}"
    read -r OSCAR_HOST
    [ -n "$OSCAR_HOST" ] || error "OSCAR_HOST is required" || return 1

    log "SSL certificate options:"
    printf "%b\n" "${YELLOW}1) Generate self-signed certificate"
    printf "%b"   "2) Use existing PEM certificate at certs/server.pem${NC}\n"
    printf "%b"   "${CYAN}Choose an option [1/2]: ${NC}"
    read -r cert_choice

    case "$cert_choice" in
        1)
            log "Generating self-signed SSL certificate for $OSCAR_HOST..."
            rm -rf certs/*
            OSCAR_HOST="$OSCAR_HOST" docker compose run --rm cert-gen || error "Failed to generate self-signed cert" || return 1
            success "Self-signed certificate generated."
            ;;
        2)
            [ -f "certs/server.pem" ] || error "certs/server.pem not found. Please place your PEM file there and rerun." || return 1
            success "Using existing certificate at certs/server.pem."
            ;;
        *)
            error "Invalid option for certificate setup." || return 1
            ;;
    esac
}

generate_nss() {
    log "Generating NSS certificate database..."
    make docker-nss || error "Failed to generate NSS cert DB" || return 1
    success "NSS certificate database created at certs/nss/"
}

start_server() {
    log "Starting Retro AIM Server with hostname $OSCAR_HOST..."
    make docker-run OSCAR_HOST="$OSCAR_HOST" || error "Failed to start server" || return 1
    success "Retro AIM Server is running."
}

final_steps() {
    printf "%b\n" "${GREEN}
Retro AIM Server setup complete!
===================================
Next Steps:

1. Copy the NSS certs from 'certs/nss/' to each AIM 6.2+ client.

2. Ensure clients can resolve '${OSCAR_HOST}' to your server IP.
   Add the following to each client's hosts file if DNS isn't used:

${YELLOW}127.0.0.1 $OSCAR_HOST${NC}

${GREEN}3. For AIM 6.x client setup instructions, see:
   https://github.com/mk6i/retro-aim-server/blob/main/docs/AIM6.md#aim-6265312-setup

Enjoy!
${NC}"
}

main() {
    check_os || return 1
    check_prereqs || return 1
    resolve_repo_root || return 1
    build_images || return 1
    setup_ssl_cert || return 1
    generate_nss || return 1
    start_server || return 1
    final_steps
}

main "$@" || true

