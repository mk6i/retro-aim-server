# Retro AIM Server Docker Compose Setup

## Prerequisites

- Docker
- Docker Compose
- Project source code

## Configuration

1. Ensure your `config/settings.env` file is properly configured.
2. Modify environment variables in `docker-compose.yml` as needed.

## Running the Server

### Start the Server
```bash
docker-compose up --build
```

### Stop the Server
```bash
docker-compose down
```

## Authentication Notes

- `DISABLE_AUTH=true` (default) allows any username/password at login
- This is a development convenience for quickly creating accounts
- To enforce authentication, set `DISABLE_AUTH=false` in `settings.env`
- When disabled, use the Management API to create and manage user accounts

## Customization

- Adjust port mappings in `docker-compose.yml`
- Add volume mounts for persistent configurations
- Modify environment variables as required
