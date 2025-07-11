# Retro AIM Server Docker Setup

This guide explains how to set up an SSL-enabled instance of Retro AIM Server using Docker.

## Prerequisites

- Git
- [Docker Desktop](https://docs.docker.com/get-started/get-docker/)
- Unix-like terminal with Makefile installed (use WSL2 for Windows)

## Quickstart script

This script will follow the steps written below and ask for user input.

```bash
curl -sSL https://github.com/mk6i/retro-aim-server/docker-setup.sh | bash
```


## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/mk6i/retro-aim-server.git
cd retro-aim-server
```

### 2. Build Docker Images

This builds Docker images for:

- Certificate generation
- SSL termination
- The Retro AIM Server runtime

```bash
make docker-images
```

### 3. Configure SSL Certificate

#### Option A: Generate a Self-Signed Certificate

If you don't have an SSL certificate, you can generate a self-signed certificate. The following creates a certificate
under `certs/server.pem`.

```bash
make docker-cert OSCAR_HOST=ras.dev
```

Replace `ras.dev` with the hostname clients will use to connect.

#### Option B: Use an Existing Certificate

If you already have an SSL certificate, place the PEM-encoded file at:

```
certs/server.pem
```

### 4. Generate NSS Certificate Database

This creates the [NSS certificate database](https://developer.mozilla.org/en-US/docs/Mozilla/Projects/NSS) in
`certs/nss/`, which must be installed on each AIM 6.2+ client.

```bash
make docker-nss
```

### 5. Start the Server

```bash
make docker-run OSCAR_HOST=ras.dev
```

Replace `ras.dev` with the hostname clients will use to connect.

### 6. Client Configuration

#### Certificate Database

Follow the [AIM 6.x client setup instructions](AIM6.md#aim-6265312-setup) to install the `certs/nss/` database on each
client.

#### Resolving Hostname

If `OSCAR_HOST` (e.g., `ras.dev`) is not a real domain with DNS configured, you'll need to add it to each client's hosts
file so clients can resolve it.

- Linux/macOS: `/etc/hosts`
- Windows: `C:\Windows\System32\drivers\etc\hosts`

Add a line like this, replacing the IP with your server's IP address:

```
127.0.0.1 ras.dev
```
