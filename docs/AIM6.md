# Windows AIM 6.x Triton Client Setup

This guide explains how to install and configure Windows AIM Triton (6.0-6.1) clients for Retro AIM Server. BUCP is
currently the only supported authentication mode.

## Installation

AIM 6.x runs well on all versions of Windows from XP-11. Unfortunately it does not run on macOS or Linux due to
incompatibility with WINE.

This guide explains how to install AIM 6.1.

1. Download AIM Triton 6.1.46.1 from [NINA wiki](https://wiki.nina.chat/wiki/Clients/AOL_Instant_Messenger#Windows).
2. Run the installation.
3. Close out of AIM.
4. Open **Task Manager** and kill the **AIM (32 bit)** process to make sure the application is
   actually terminated.

## Configuration

Unlike previous versions of AIM that allow you to enter server configuration in the application settings, AIM 6.x
settings must be modified in configuration files.

First change the authentication mode from Kerberos to BUCP.

1. Press Start, type Notepad, right-click it, and choose Run as administrator.
2. Inside Notepad, go to **File** → **Open**.
3. Above the **Open** button, select **All Files**
4. Navigate to `C:\Program Files (x86)\AIM6\services\im\ver1_14_9_1`.
5. Select `serviceManifest.xml`, and click **Open**.
6. Replace the value of the `<preferenceDefault>` with `BUCP`.
   ```diff
   -<preferenceDefault key="aol.im.connect.mode" scope="identity" type="string">AAS</preferenceDefault>
   +<preferenceDefault key="aol.im.connect.mode" scope="identity" type="string">BUCP</preferenceDefault>
   -<preferenceDefault key="aol.im.connect.mode2" scope="identity" type="string">AAAS</preferenceDefault>
   +<preferenceDefault key="aol.im.connect.mode2" scope="identity" type="string">BUCP</preferenceDefault>
   ```
7. Click **File** → **Save**.

Next change the server hostname configuration.

1. Still inside Notepad, go to **File** → **Open**.
2. Above the **Open** button, select **All Files**
3. Navigate to `C:\Program Files (x86)\AIM6\services\imApp\ver6_1_46_1`.
4. Select `serviceManifest.xml`, and click **Open**.
5. Change the <preferenceDefault> value for key `aol.aimcc.connect.host.address` from `login.oscar.aol.com` to the value of `OSCAR_HOST` found in `config/settings`.
   ```diff
   -<preferenceDefault key="aol.aimcc.connect.host.address" scope="identity" type="string">login.oscar.aol.com</preferenceDefault>
   +<preferenceDefault key="aol.aimcc.connect.host.address" scope="identity" type="string">127.0.0.1</preferenceDefault>
   ```
7. Click **File** → **Save**.
