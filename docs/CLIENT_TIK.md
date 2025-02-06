# TiK Client Setup

This guide explains how to install and configure **TiK** for Retro AIM Server.

 <p align="center">
    <img width="400" alt="screenshot of TiK sign-on window" src="https://github.com/user-attachments/assets/30f6b91b-cfe9-4749-a8f6-8f089ff24125">
 </p>

[TiK](https://en.wikipedia.org/wiki/TiK) is an open source instant messenger developed by AOL in the late 1990s. It
communicates with AIM servers using TOC, a text-based alternative to the OSCAR protocol.

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

   Download mk6i's [fork of TiK v0.90](https://github.com/mk6i/tik/archive/refs/heads/main.zip), which includes
   compatibility fixes for modern systems.

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

   In the file listing, double-click **tik.tcl** to launch TiK. A setup window appears.

5. **Set Configuration**

   Click the **Advanced** button in the setup prompt. This closes the setup window.

   <p align="center">
      <img width="400" alt="screenshot of TiK setup prompt" src="https://github.com/user-attachments/assets/5b29075b-8d21-42e8-8806-71009660a628">
   </p>

6. **Configure TOC Hostname**

   From the login window, click the **Configure** button, which brings up the connection config window.

   <p align="center">
      <img width="400" alt="screenshot of TiK connection config window" src="https://github.com/user-attachments/assets/590b8100-0f7a-45d0-b210-bc63202beb0e">
   </p>

   Enter the hostname and port of the TOC server you want to connect and click **OK**.

   If running your own server, set the values that correspond to `TOC_HOST` and `TOC_PORT` in the **settings.env**
   config file.

7. **Sign On**

   Relaunch TiK and sign in!

### Linux

1. **Install Tcl/Tk 8.x**

   Open a terminal and install Tcl/Tk 8.x. The following example works for Ubuntu. Install the analogous packages for
   your distro of choice.

    ```shell
   apt install tcl tk
   ```

2. **Download TiK**

   Download mk6i's [fork of TiK v0.90](https://github.com/mk6i/tik/archive/refs/heads/main.tar.gz), which includes
   compatibility fixes for modern systems.

   In the terminal, extract the downloaded archive and **cd** into the **tik-main** directory.

   ```shell
   tar xvf /path/to/downloads/tik-main.tar.gz
   cd /path/to/downloads/tik-main
   ```

3. **Launch TiK**

   Enter the following command in the terminal to launch TiK. A setup window appears.

   ```shell
   ./tik.tcl
   ```

4. **Set Configuration**

   Click the **Advanced** button in the setup prompt. This closes the setup window.

   <p align="center">
      <img width="400" alt="screenshot of TiK setup prompt" src="https://github.com/user-attachments/assets/fbae111b-e2ee-4a50-a298-256d22a69422">
   </p>

5. **Configure TOC Hostname**

   From the login window, click the **Configure** button, which brings up the connection config window.

   <p align="center">
      <img width="400" alt="screenshot of TiK connection config window" src="https://github.com/user-attachments/assets/8bd7d6cd-1403-4656-8acc-e36ff8070b61">
   </p>

   Enter the hostname and port of the TOC server you want to connect and click **OK**.

   If running your own server, set the values that correspond to `TOC_HOST` and `TOC_PORT` in the **settings.env**
   config file.

6. **Sign On**

   Now enter your **Screen Name** and **Password** and sign on to AIM!

### macOS (Intel & Apple Silicon)

1. **Install Tcl/Tk 8.x**

   Open a terminal and install Tcl/Tk 8.x using [Homebrew](https://brew.sh/).

    ```shell
   brew install tcl-tk@8
   ```

2. **Verify Tcl/Tk Version**

   By default, macOS includes a Tcl/Tk version that does not support TiK. Verify that version 8.x installed in the
   previous step is in your PATH. Run the following command in a new terminal:

   ```shell
   echo 'puts $tcl_version' | tclsh
   ```

   If the reported version **is not 8.x**, run `brew link tcl-tk@8` or `brew doctor` to fix the installation.

3. **Download TiK**

   Download mk6i's [fork of TiK v0.90](https://github.com/mk6i/tik/archive/refs/heads/main.zip), which includes
   compatibility fixes for modern systems.

   In the terminal, extract the downloaded archive and **cd** into the **tik-main** directory.

   ```shell
   unzip /path/to/downloads/tik-main.zip
   cd /path/to/downloads/tik-main
   ```

4. **Launch TiK**

   Enter the following command in the terminal to launch TiK. A setup window appears.

   ```shell
   ./tik.tcl
   ```

5. **Set Configuration**

   Click the **Advanced** button in the setup prompt. This closes the setup window.

   <p align="center">
      <img width="400" alt="screenshot of TiK setup prompt" src="https://github.com/user-attachments/assets/13956991-235b-422e-8b4a-f05a5974637d">
   </p>

6. **Configure TOC Hostname**

   From the login window, click the **Configure** button, which brings up the connection config window.

   <p align="center">
      <img width="400" alt="screenshot of TiK connection config window" src="https://github.com/user-attachments/assets/aa89836e-c0a9-40a4-9fb7-1b809bb55ffd">
   </p>

   Enter the hostname and port of the TOC server you want to connect and click **OK**.

   If running your own server, set the values that correspond to `TOC_HOST` and `TOC_PORT` in the **settings.env**
   config file.

7. **Sign On**

   Now enter your **Screen Name** and **Password** and sign on to AIM!
