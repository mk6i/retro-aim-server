# Development: How to Run/Compile Retro AIM Server

This document explains how to Build Retro AIM Server.

## Dependencies

A C compiler is required in order to build the sqlite dependency.

**macOS**

> If you have git, this is likely already set up on your machine.

```shell
xcode-select --install
```

**Linux (Ubuntu)**

```shell
sudo apt install build-essential
```

Retro AIM Server requires [go 1.21](https://go.dev/) or newer to run.

## Run the Server

To run Retro AIM Server without building a binary, run the following script. The default settings can be modified
in `config/settings.env`.

```shell
scripts/run_dev.sh
```

## Build the Server

```shell
go build -o retro_aim_server ./cmd/server
```

To run:

```shell
source config/settings.env
./retro_aim_server
```
