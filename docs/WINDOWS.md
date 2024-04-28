# Retro AIM Server Quickstart for Windows 10/11

This guide explains how to download, configure and run Retro AIM Server on Windows 10/11.

1. **Download Retro AIM Server**

   Grab the latest Windows release from the [Releases page](https://github.com/mk6i/retro-aim-server/releases).

   Because the Retro AIM Server `.exe` has not been blessed by Microsoft, browsers such as Chrome may think it's a
   "suspicious" file and block the download, in which case you need to explicitly opt in to downloading the untrusted
   file.

    <p align="center">
      <img alt="screenshot of a chrome prompt showing a blocked download" src="https://github.com/mk6i/retro-aim-server/assets/2894330/6bf2fd79-0a42-48b2-a695-d777259a3603">
    </p>

   > While the binaries are 100% safe, you can avoid the security concern by [building the application yourself](./BUILD.md).
   We do not provide signed binaries because of the undue cost and complexity.

   Once downloaded, extract the `.zip` archive, which contains the application and a configuration file `settings.bat`.

2. **Configure Server Address**

   Open `settings.bat` (right-click, `edit in notepad`) and set `OSCAR_HOST` to a hostname that AIM clients can connect
   to. The default setting is `127.0.0.1`, which is enough to connect clients on the same PC.

   In order to connect AIM clients on your LAN (including VMs with bridged networking), you can find the appropriate IP
   address by running `ipconfig` from the Command Prompt.

3. **Start the Application**

   Open `run.cmd` to launch Retro AIM Server.

   Because Retro AIM Server has not been blessed by Microsoft, Windows will flag the application as a security risk the
   first time you run it. You'll be presented with a `Microsoft Defender SmartScreen` warning prompt that gives you the
   option to run the blocked application.

   To proceed, click `More Options`, then `Run anyway`.

    <p align="center">
      <img alt="of screenshot microsoft defender smartscreen prompt" src="https://github.com/mk6i/retro-aim-server/assets/2894330/9ab0966b-d5dd-4b70-ba16-483e5c458f89">
      <img alt="of screenshot microsoft defender smartscreen prompt" src="https://github.com/mk6i/retro-aim-server/assets/2894330/5d4106c6-0ce6-4d4f-9260-e9bbb777c770">
    </p>

   > While the binaries are 100% safe, you can avoid the security concern by [building the application yourself](./BUILD.md).
   We do not provide signed binaries because of the undue cost and complexity.

   Retro AIM Server will open in a Command Prompt, ready to accept AIM client connections.

4. **Test**

   To do a quick sanity check, start an AIM client, sign in to the server, and send yourself an instant message.
   Configure the AIM client to connect to the host set in `OSCAR_HOST` in `settings.bat`. (If you didn't change the
   config, the address is `127.0.0.1`.)

   See the [Client Configuration Guide](./CLIENT.md) for more detail on setting up the AIM client.

   By default, you can enter *any* screen name and password at the AIM sign-in screen to auto-create an account.

   > Account auto-creation is meant to be a convenience feature for local development. In a production deployment, you
   should set `DISABLE_AUTH=false` in `settings.bat` to enforce account authentication. User accounts can be created via
   the [Management API](../README.md#-management-api).