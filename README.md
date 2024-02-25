<p align="center">
  <img src="https://github.com/mk6i/retro-aim-server/assets/2894330/adff6b45-fcae-400c-8876-52e891a36ee0" width="300">
</p>

<p align="center">
  <a href="https://codecov.io/github/mk6i/retro-aim-server">
    <img src="https://codecov.io/github/mk6i/retro-aim-server/graph/badge.svg?token=MATKPP77JT" alt="codecov">
  </a>
</p>

Retro AIM Server is a server implementation of the OSCAR protocol that supports AIM versions 5.0-5.9.

This project is currently under heavy development. Retro AIM Server supports/will support the following features:

- [x] Instant Messaging
- [x] Buddy List
- [x] Warning
- [x] Away Messages
- [x] User Profiles
- [x] Chat Rooms
- [x] Visibility Toggle
- [x] User Blocking
- [ ] Buddy Icons
- [ ] User Directory

## Quickstart

### Dependencies

A C compiler is required in order to build the sqlite dependency.

**MacOS**

> If you have git, this is likely already set up on your machine.

```shell
xcode-select --install
```

**Linux (Ubuntu)**

```shell
sudo apt install build-essential
```

Retro AIM Server requires [go 1.21](https://go.dev/) or newer to run.

### Run the Server

Start Retro AIM Server with the following command. The default settings can be modified in `config/settings.env`.

```shell
scripts/run_dev.sh
```

### Configure AIM Client

Download Windows AIM ([v5.1.3036 recommended](https://archive.org/details/aim513036)) and install on Windows 10/11 using
[Windows XP compatibility mode](https://support.microsoft.com/en-us/windows/make-older-apps-or-programs-compatible-with-windows-783d6dd7-b439-bdb0-0490-54eea0f45938)
or MacOS/Linux via [Wine](https://www.winehq.org/). 

Once installed, configure the AIM client to connect to Retro AIM Server as
follows:

1. At the sign-on screen, click `Setup`.
2. Under the `Sign On/Off` category, click `Connection`.
3. In the `Server  > Host` field, enter the value of `OSCAR_HOST` found in `config/settings.env`.
4. In the `Server > Port` field, enter the value of `BOS_PORT` found in `config/settings.env`.

Apply the settings and sign on to AIM. By default, you can sign on with any screen name/password without first
registering (see `DISABLE_AUTH` in `config/settings.env` for more details).

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
