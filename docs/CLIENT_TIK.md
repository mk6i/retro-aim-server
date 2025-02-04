# TiK Client Setup

This guide explains how to install and configure **TiK** for Retro AIM Server.

 <p align="center">
    <img width="400" alt="screenshot of TiK sign-on window" src="https://github.com/user-attachments/assets/7842b8ef-b2c2-482f-841a-c124669180e3">
 </p>

[TiK](https://en.wikipedia.org/wiki/TiK) is an open source instant messenger developed by AOL in the late 1990s. It
communicates using TOC, a text-based alternative to the OSCAR protocol.

Installation guides are available for the following operating systems:

* [Windows](#windows)
* [Linux](#linux)
* [macOS (Intel & Apple Silicon)](#macos-intel--apple-silicon)

## Installation

### Windows

1. **Install Tcl/Tk 8.x**

   Download and install the latest 8.x version of the **Magicsplat Tcl/Tk distribution**
   from [Sourceforge](https://sourceforge.net/projects/magicsplat/files/magicsplat-tcl/tcl-8.6.16-installer-1.16.0-x64.msi/download).

2. **Download TiK**

   Download mk6i's [fork of TiK .90](https://github.com/mk6i/tik/archive/refs/heads/main.zip). The fork contains fixes
   that smooth out the TiK setup experience on modern systems.

3. **Create File Association**

   The following instructions show how to make **.tcl** files open with the **wish** interpreter.

   - Extract the downloaded archive in **File Explorer** and open the TiK folder.
   - In the file listing, right-click **tik.tcl** → **Open with...** 
     - **Windows10**: → Click **More apps** → Click **Look for another app on this PC**.
     - **Windows11**: → Click **Choose an app on your PC**.
   - In the **Open with...** dialog, navigate to the following path:
      ```
      %USERPROFILE%\AppData\Local\Apps\Tcl86\bin
      ```
   - Select **wish.exe** and click **Open**.
   - **Windows11**: Click the **Always** button.

4. **Launch TiK**

   In the file listing, double-click **tik.tcl** to launch TiK.

5. **Configure TOC Hostname**

   From the login window, click the **Configure** button, which brings up the connection config window.

   <p align="center">
      <img width="400" alt="screenshot of TiK connection config window" src="https://github.com/user-attachments/assets/aa89836e-c0a9-40a4-9fb7-1b809bb55ffd">
   </p>

   Enter the hostname and port of the TOC server you want to connect to in the **TOC Host** and **TOC Port** fields and
   click *OK*.

   If running your own server, set the values that correspond to `TOC_HOST` and `TOC_PORT` in the `settings.env` config
   file.

6. **Sign On**

   Relaunch TiK and sign in!

### Linux

1. **Install Tcl/Tk 8.x**

   Open a terminal and install Tcl/Tk 8.x. The following example works for Ubuntu. Install the analogous packages for
   your distro of choice.

    ```shell
   apt install tcl tk
   ```

2. **Download TiK**

   Download mk6i's [fork of TiK .90](https://github.com/mk6i/tik/archive/refs/heads/main.tar.gz). The fork contains
   fixes
   that smooth out the TiK setup experience on modern systems.

3. **Launch TiK**

   Return to the terminal and extract the archive downloaded in the previous step.

   ```shell
   unzip ~/Downloads/tik-main.zip
   cd ~/Downloads/tik-main
   ```

   Then launch TiK...

   ```shell
   ./tik.tcl
   ```

   A setup prompt and login window appear. Click the **Advanced** button in the setup prompt, which closes the setup
   window.

   <p align="center">
      <img width="400" alt="screenshot of TiK setup prompt" src="https://github.com/user-attachments/assets/4c709f1b-7299-4567-bfcd-afe715e961b5">
   </p>

4. **Configure TOC Hostname**

   From the login window, click the **Configure** button, which brings up the connection config window.

   <p align="center">
      <img width="400" alt="screenshot of TiK connection config window" src="https://github.com/user-attachments/assets/aa89836e-c0a9-40a4-9fb7-1b809bb55ffd">
   </p>

   Enter the hostname and port of the TOC server you want to connect to in the **TOC Host** and **TOC Port** fields and
   click *OK*.

   If running your own server, set the values that correspond to `TOC_HOST` and `TOC_PORT` in the `settings.env` config
   file.

5. **Sign On**

   Now enter your **Screen Name** and **Password** and sign on to AIM!

### macOS (Intel & Apple Silicon)

1. **Install Tcl/Tk 8.x**

   Open a terminal and install Tcl/Tk 8.x using [Homebrew](https://brew.sh/).

    ```shell
   brew install tcl-tk@8
   ```

2. **Verify Tcl/Tk Version**

   macOS comes by default with a version of Tcl/Tk that does not support TiK. Verify that version 8.x installed in
   the previous step is in your PATH. Run the following command in a new terminal:

   ```shell
   echo 'puts $tcl_version' | tclsh
   ```

   If the reported version **is not 8.x**, try running `brew link tcl-tk@8` or `brew doctor` to fix the installation.

3. **Download TiK**

   Download mk6i's [fork of TiK .90](https://github.com/mk6i/tik/archive/refs/heads/main.zip). The fork contains fixes
   that smooth out the TiK setup experience on modern systems.

4. **Launch TiK**

   Return to the terminal and extract the archive downloaded in the previous step.

   ```shell
   unzip ~/Downloads/tik-main.zip
   cd ~/Downloads/tik-main
   ```

   Then launch TiK...

   ```shell
   ./tik.tcl
   ```

   A setup prompt and login window appear. Click the **Advanced** button in the setup prompt, which closes the setup
   window.

   <p align="center">
      <img width="400" alt="screenshot of TiK setup prompt" src="https://github.com/user-attachments/assets/4c709f1b-7299-4567-bfcd-afe715e961b5">
   </p>

5. **Configure TOC Hostname**

   From the login window, click the **Configure** button, which brings up the connection config window.

   <p align="center">
      <img width="400" alt="screenshot of TiK connection config window" src="https://github.com/user-attachments/assets/aa89836e-c0a9-40a4-9fb7-1b809bb55ffd">
   </p>

   Enter the hostname and port of the TOC server you want to connect to in the **TOC Host** and **TOC Port** fields and
   click *OK*.

   If running your own server, set the values that correspond to `TOC_HOST` and `TOC_PORT` in the `settings.env` config
   file.

6. **Sign On**

   Now enter your **Screen Name** and **Password** and sign on to AIM!
