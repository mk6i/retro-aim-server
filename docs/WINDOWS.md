# Retro AIM Server Quickstart for Windows 10/11

This guide explains how to download, configure and run Retro AIM Server on Windows 10/11.

1. **Download Retro AIM Server**

   Download the latest Windows release from the [Releases page](https://github.com/mk6i/retro-aim-server/releases) and
   extract the `.zip` archive, which contains the application and a configuration file `settings.env`.

2. **Configure Server Address**

   Open `settings.env` (right-click, `edit in notepad`) and set `OSCAR_HOST` to a hostname that AIM clients can connect
   to. The default setting is `127.0.0.1`, which is enough to connect clients on the same PC.

   In order to connect AIM clients on your LAN (including VMs with bridged networking), you can find the appropriate IP
   address by running `ipconfig` from the Command Prompt.

3. **Start the Application**

   Launch `retro-aim-server.exe` to start Retro AIM Server.

   Because Retro AIM Server has not built up enough reputation with Microsoft, Windows will flag the application as a
   security risk the first time you run it. You'll be presented with a `Microsoft Defender SmartScreen` warning prompt
   that gives you the option to run the blocked application.

   To proceed, click `More Options`, then `Run anyway`.

    <p align="center">
      <img alt="screenshot of microsoft defender smartscreen prompt" src="https://github.com/mk6i/retro-aim-server/assets/2894330/9ab0966b-d5dd-4b70-ba16-483e5c458f89">
      <img alt="screenshot of microsoft defender smartscreen prompt" src="https://github.com/mk6i/retro-aim-server/assets/2894330/5d4106c6-0ce6-4d4f-9260-e9bbb777c770">
    </p>

   Click `Allow` if you get a Windows Defender Firewall alert.

    <p align="center">
      <img alt="screenshot of microsoft defender firewall alert" src="https://github.com/user-attachments/assets/9ec6cbc4-5445-43bd-a64e-512fd15f8f0b">
    </p>

   Retro AIM Server will open in a terminal, ready to accept AIM client connections.

4. **Test**

   To do a quick sanity check, start an AIM client, sign in to the server, and send yourself an instant message.
   Configure the AIM client to connect to the host set in `OSCAR_HOST` in `settings.env`. (If you didn't change the
   config, the address is `127.0.0.1`.)

   See the [Client Configuration Guide](./CLIENT.md) for more detail on setting up the AIM client.

   By default, you can enter *any* screen name and password at the AIM sign-in screen to auto-create an account.

   > Account auto-creation is meant to be a convenience feature for local development. In a production deployment, you
   should set `DISABLE_AUTH=false` in `settings.env` to enforce account authentication. User accounts can be created via
   the [Management API](../README.md#-management-api).
