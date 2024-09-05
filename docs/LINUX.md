# Retro AIM Server Quickstart for Linux (x86_64)

This guide explains how to download, configure and run Retro AIM Server on Linux (x86_64).

1. **Download Retro AIM Server**

   Grab the latest Linux release from the [releases page](https://github.com/mk6i/retro-aim-server/releases) and extract
   the archive. The extracted folder contains the application and a configuration file `settings.env`.

2. **Configure Server Address**

   Set `OSCAR_HOST` in `settings.env` to a hostname that AIM clients can connect to. The default setting is `127.0.0.1`,
   which is enough to connect clients on the same PC.

   In order to connect AIM clients on your LAN (including VMs with bridged networking), you can find the appropriate IP
   address by running `ifconfig` from the terminal

3. **Start the Application**

   Run the following command to launch Retro AIM Server:

   ```shell
   ./retro_aim_server
   ```

   Retro AIM Server will run in the terminal, ready to accept AIM client connections.

4. **Configure AIM Clients**

   To do a quick sanity check, start an AIM client, sign in to the server, and send yourself an instant message.
   Configure the AIM client to connect to the host set in `OSCAR_HOST` in `settings.env`. (If you didn't change the
   config, the address is `127.0.0.1`.)

   See the [Client Configuration Guide](./CLIENT.md) for more detail on setting up the AIM client.

   By default, you can enter *any* screen name and password at the AIM sign-in screen to auto-create an account.

   > Account auto-creation is meant to be a convenience feature for local development. In a production deployment, you
   should set `DISABLE_AUTH=false` in `settings.env` to enforce account authentication. User accounts can be created via
   the [Management API](../README.md#-management-api).