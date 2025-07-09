# Windows AIM 6.x Client Setup

This guide explains how to install and configure **AIM 6.x** for use with **Retro AIM Server**.

<p align="center">
   <img alt="screenshot of AIM sign-on screen" src="https://github.com/user-attachments/assets/057c72fe-3d60-4dad-a602-8ff95c4fcbe1">
</p>

Installation guides are available for the following versions:

- [AIM 6.0-6.1](#aim-60-61-setup) (BUCP auth)
- [AIM 6.2-6.5](#aim-6265-setup) (Kerberos auth)

## AIM 6.0-6.1 Setup

### Installation

1. Download AIM 6.x (recommended **AIM 6.1.46.1**) from
   the [NINA wiki](https://wiki.nina.chat/wiki/Clients/AOL_Instant_Messenger#Windows).
2. Run the installer and complete the installation.
3. Close the AIM application.
4. Open **Task Manager** and end the **AIM (32 bit)** process if it's still running.

### Configure Authentication Mode

AIM 6.x does not expose server settings via the UI. You'll need to edit configuration files manually.

To switch from the default Kerberos-based auth (AAM/AAMUAS) to BUCP:

1. Open **Notepad as Administrator** (Start → type "Notepad" → right-click → **Run as Administrator**).
2. In Notepad, go to **File → Open**.
3. Navigate to:
   ```
   C:\Program Files (x86)\AIM6\services\im\ver1_14_9_1
   ```
4. Change the file filter to **All Files**.
5. Open `serviceManifest.xml`.
6. Locate the `aol.im.connect.mode` and `aol.im.connect.mode2` preferences and change them from `AAM` and `AAMUAS` to
   `BUCP`:

   ```diff
   -<preferenceDefault key="aol.im.connect.mode" scope="identity" type="string">AAM</preferenceDefault>
   +<preferenceDefault key="aol.im.connect.mode" scope="identity" type="string">BUCP</preferenceDefault>
   -<preferenceDefault key="aol.im.connect.mode2" scope="identity" type="string">AAMUAS</preferenceDefault>
   +<preferenceDefault key="aol.im.connect.mode2" scope="identity" type="string">BUCP</preferenceDefault>
   ```

7. Save the file.

### Configure Server Hostname

To point the client to your Retro AIM Server:

1. In Notepad, go to **File → Open** again.
2. Navigate to:
   ```
   C:\Program Files (x86)\AIM6\services\imApp\ver6_1_46_1
   ```
3. Set the file filter to **All Files**.
4. Open `serviceManifest.xml`.
5. Find the `aol.aimcc.connect.host.address` preference and update it to match your `OSCAR_HOST` Retro AIM Server
   config:

   ```diff
   -<preferenceDefault key="aol.aimcc.connect.host.address" scope="identity" type="string">login.oscar.aol.com</preferenceDefault>
   +<preferenceDefault key="aol.aimcc.connect.host.address" scope="identity" type="string">127.0.0.1</preferenceDefault>
   ```

6. Save the file.

## AIM 6.2–6.5.3.12 Setup

### Installation

1. Download an AIM 6.2–6.5 client (recommended **AIM 6.5.3.12**) from the  
   [NINA wiki](https://wiki.nina.chat/wiki/Clients/AOL_Instant_Messenger#Windows).
2. Run the installer and complete the installation.
3. Close the AIM application.
4. Open **Task Manager** and end the **AIM (32 bit)** process if it's still running.

### Install Certificate Database

Follow [these steps](CLIENT.md) to generate a certificate database or request one from the server operator, if
applicable.

Copy the following files to `%APPDATA%\acccore\nss`:

- `cert8.db`
- `key3.db`
- `secmod.db`

### Configure Server Hostname

Tell AIM where to connect:

1. Start AIM.
2. Open **Settings**, then go to the **Connection** tab.
3. In the **Host** field, enter the domain name that matches the certificate's Common Name (CN).
4. In the **Port** field, enter `443`.
5. Click **Save**, then sign in!

## Enable Legacy JavaScript Engine (Windows 11 24H2+ Only)

AIM 6.x's frontend breaks under the new JavaScript engine introduced in Windows 11 24H2. A workaround described by
[axelsw.it](https://www.axelsw.it/pwiki/index.php/JScript_Windows11) forces Windows to use an older JavaScript engine
compatible with AIM 6.x.

> ⚠️ Downgrading the JavaScript engine is generally a bad idea, as it may expose your system to vulnerabilities fixed in
> newer engines.
> **Proceed at your own risk!**

To implement the workaround, create a `.reg` file with the following content. Double-click the file in Windows Explorer
to apply the change.

```
Windows Registry Editor Version 5.00

[HKEY_CURRENT_USER\Software\Policies\Microsoft\Internet Explorer\Main]
"JScriptReplacement"=dword:00000000
```
