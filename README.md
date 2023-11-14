# Retro AIM Server

Retro AIM Server is a server implementation of the OSCAR protocol that supports AIM versions 5.0-5.9.

## How to run

Retro AIM Server requires [go 1.21](https://go.dev/).

### Configuration

Server configuration is set through environment variables. The following are the most useful configs:

| Env Variable   | Description                                                                                              |
|----------------|----------------------------------------------------------------------------------------------------------|
| `OSCAR_HOST`   | The hostname that the server should bind to. If exposing to the internet, use the public IP.             |
| `DISABLE_AUTH` | If true, auto-create screen names at login and skip the password check. Useful for development purposes. |
| `DB_PATH`      | The path to the SQLite database.                                                                         |
| `LOG_LEVEL`    | Set logging granularity. Possible values: `trace`, `debug`, `info`, `warn`, `error`                      |

### Starting Up

```shell
DISABLE_AUTH=true \
OSCAR_HOST=192.168.64.1 \
DB_PATH=./aim.db \
go run ./cmd/main.go
```

### User Management

User management is done through a REST API.

#### List Users

```curl
curl http://localhost:8080/user
```

#### Create Users

```curl
curl -d'{"screen_name":"myScreenName", "password":"thepassword"}' http://localhost:8080/user
```
