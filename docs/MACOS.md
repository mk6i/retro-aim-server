# Retro AIM Server Quickstart for macOS (Intel and Apple Silicon)

This guide explains how to download, configure and run Retro AIM Server on macOS (Intel and Apple Silicon).

1. **Download Retro AIM Server**

   Grab the latest macOS release from the [Releases page](https://github.com/mk6i/retro-aim-server/releases) for your
   platform (Intel or Apple Silicon).

   Because the Retro AIM Server `.app` has not been blessed by Apple, browsers such as Chrome may think it's a
   "suspicious" file and block the download, in which case you need to explicitly opt in to downloading the untrusted
   file.

    <p align="center">
      <img alt="screenshot of a chrome prompt showing a blocked download" src="https://github.com/mk6i/retro-aim-server/assets/2894330/90af40bd-262d-4e0f-a769-06943c7acd18">
    </p>

   > While the binaries are 100% safe, you can avoid the security concern by [building the application yourself](./BUILD.md).
   We do not provide signed binaries because of the undue cost and complexity.

   Once downloaded, extract the `.zip` archive, which contains the application and a configuration file `settings.env`.

2. **Open Terminal**

   Open a terminal and navigate to the extracted directory. This terminal will be used for the remaining steps.

   ```shell
   cd ~/Downloads/retro_aim_server.0.1.0.macos.intel_x86_64/
   ```

3. **Remove Quarantine**

   Because the Retro AIM Server `.app` has not been blessed by Apple, macOS will quarantine the application. To proceed,
   remove the quarantine flag from the `.app`. In the same terminal, run following command:

   ```shell
   sudo xattr -d com.apple.quarantine ./bin/retro_aim_server
   ```

   > While the binaries are 100% safe, you can avoid the security concern by [building the application yourself](./BUILD.md).
   We do not provide signed binaries because of the undue cost and complexity.

4. **Configure Server Address**

   Set `OSCAR_HOST` in `settings.env` to a hostname that AIM clients can connect to. The default setting is `127.0.0.1`,
   which is enough to connect clients on the same machine.

   In order to connect AIM clients on your LAN (including VMs with bridged networking), you can find the appropriate IP
   address by running the following command in the terminal:

   ```shell
   osascript -e "IPv4 address of (system info)"
   ```

5. **Start the Application**

   Run the following command to launch Retro AIM Server:

   ```shell
   ./run.sh
   ```

   Retro AIM Server will run in the terminal, ready to accept AIM client connections.

6. **Test**

   To do a quick sanity check, start an AIM client, sign in to the server, and send yourself an instant message.
   Configure the AIM client to connect to the host set in `OSCAR_HOST` in `settings.env`. (If you didn't change the
   config, the address is `127.0.0.1`.)

   See the [Client Configuration Guide](./CLIENT.md) for more detail on setting up the AIM client.

   By default, you can enter *any* screen name and password at the AIM sign-in screen to auto-create an account.

   > Account auto-creation is meant to be a convenience feature for local development. In a production deployment, you
   should set `DISABLE_AUTH=false` in `settings.env` to enforce account authentication. User accounts can be created via
   the [Management API](../README.md#-management-api).