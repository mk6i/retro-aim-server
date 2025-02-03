# TiK Client Setup

This guide explains how to install and configure **TiK** for Retro AIM Server.

 <p align="center">
    <img width="400" alt="screenshot of TiK sign-on window" src="https://github.com/user-attachments/assets/7842b8ef-b2c2-482f-841a-c124669180e3">
 </p>

[TiK](https://en.wikipedia.org/wiki/TiK) is an open source instant messenger developed by AOL in the late 1990s. It
communicates using TOC, a text-based alternative to the OSCAR protocol.

In order to run TiK, you'll need to install the [Tcl/Tk runtime](https://www.tcl-lang.org/).

Installation guides are available for the following operating systems:

* [Windows](#windows)
* [Linux](#linux)
* [macOS (Intel & Apple Silicon)](#macos-intel--apple-silicon)

## Installation

### Windows

1. **Download TiK**

   Download TiK from [Sourceforge](https://sourceforge.net/projects/tik/files/tik/) and extract the archive.

    - **[v0.75](https://sourceforge.net/projects/tik/files/tik/0.75/)** → Last official version by AOL.
    - **[v0.90](https://sourceforge.net/projects/tik/files/tik/0.90/)** → Last community release (**Recommended**).

2. **Install Tcl/Tk 8.x**

   Download and install the latest 8.x version of the **Magicsplat Tcl/Tk distribution**
   from [Sourceforge](https://sourceforge.net/projects/magicsplat/files/magicsplat-tcl/).

3. **Create Config Directory**

    - Open `%USERPROFILE%\Documents` in **File Explorer**.
    - Create a folder called `tik`.

4. **Create File Association**

   Ensure `.tcl` files open with the `wish` interpreter from Tcl/Tk.

    - Open the folder extracted from the TiK archive in Step 1.
    - Right-click **tik.tcl** → **Open with...** → Click **More apps** → Click **Look for another app on this PC**.
    - In the **Open with...** dialog, navigate to the following path:
       ```
       %USERPROFILE%\AppData\Local\Apps\Tcl86\bin
       ```
    - Select **wish.exe** and click **Open**.

5. **Launch TiK**

   Double-click `tik.tcl` to launch TiK and immediately **close the application**.

6. **Configure Hostname**

   > This step is a workaround to a bug that makes TiK unable to persist connection settings to the config file.

    - Open `%USERPROFILE%\Documents\tik` in **File Explorer**.
    - Right-click **tikrc** → **Open with** → **Notepad**.
    - **Uncomment** the line beginning with `#set TOC(production,host)`.
    - Change the default hostname value `10.10.10.10` to the AIM server hostname. (If running your own Retro AIM Server,
      the hostname must correspond to the `TOC_HOST` env variable.)

   ```diff
   - #set TOC(production,host) "10.10.10.10"  ;# IP of toc.oscar.aol.com
   + set TOC(production,host) "127.0.0.1"
   ```

7. **Sign on**

   Relaunch TiK and sign in!

### Linux

1. **Download TiK**

   Download TiK from [Sourceforge](https://sourceforge.net/projects/tik/files/tik/) and extract the archive.

    - **[v0.75](https://sourceforge.net/projects/tik/files/tik/0.75/)** → Last official version by AOL.
    - **[v0.90](https://sourceforge.net/projects/tik/files/tik/0.90/)** → Last community release (**Recommended**).

2. **Install Tcl/Tk 8.x**

   Open a terminal and install Tcl/Tk 8.x. The following example works for Ubuntu. Install the analogous packages for
   your distro of choice.

    ```shell
   apt install tcl tk
   ```

3. **Create Config Directory**

   > Steps 3-4 are a workaround to a bug that causes the first-time setup wizard to freeze.

   ```shell
   mkdir ~/.tik
   ```

4. **Launch TiK**

   In the terminal, navigate to the extracted TiK archive. Launch TiK and immediately **close the application**.

   ```shell
   ./tik.tcl
   ```

5. **Configure TOC Hostname**

   > This step is a workaround to a bug that makes TiK unable to persist connection settings to the config file.

    - Open `~/.tik/tikrc` in your favorite editor.
    - **Uncomment** the line beginning with `#set TOC(production,host)`.
    - Change the default hostname value `10.10.10.10` to the AIM server hostname. (If running your own Retro AIM Server,
      the hostname must correspond to the `TOC_HOST` env variable.)

   ```diff
   - #set TOC(production,host) "10.10.10.10"  ;# IP of toc.oscar.aol.com
   + set TOC(production,host) "127.0.0.1"
   ```

6. **Sign on**

   Relaunch TiK and sign in!

### macOS (Intel & Apple Silicon)

1. **Download TiK**

   Download TiK from [Sourceforge](https://sourceforge.net/projects/tik/files/tik/) and extract the archive.

    - **[v0.75](https://sourceforge.net/projects/tik/files/tik/0.75/)** → Last official version by AOL.
    - **[v0.90](https://sourceforge.net/projects/tik/files/tik/0.90/)** → Last community release (**Recommended**).

2. **Install Tcl/Tk 8.x**

   Open a terminal and install Tcl/Tk 8.x using [Homebrew](https://brew.sh/).

    ```shell
   brew install tcl-tk@8
   ```

3. **Verify Tcl/Tk Version**

   macOS comes by default with a version of Tcl/Tk that does not support TiK. Verify that version 8.x installed in
   the previous step is in your PATH. Run the following command in a new terminal:

   ```shell
   echo 'puts $tcl_version' | tclsh
   ```

   If the reported version **is not 8.x**, try running `brew link tcl-tk@8` or `brew doctor` to fix the installation.

4. **Create Config Directory**

   > Steps 4-5 are a workaround to a bug that causes the first-time setup wizard to freeze.

   ```shell
   mkdir ~/.tik
   ```

5. **Launch TiK**

   In the terminal, navigate to the extracted TiK archive. Launch TiK and immediately **close the application**.

   ```shell
   ./tik.tcl
   ```

6. **Configure TOC Hostname**

   > This step is a workaround to a bug that makes TiK unable to persist connection settings to the config file.

    - Open `~/.tik/tikrc` in your favorite editor.
    - **Uncomment** the line beginning with `#set TOC(production,host)`.
    - Change the default hostname value `10.10.10.10` to the AIM server hostname. (If running your own Retro AIM Server,
      the hostname must correspond to the `TOC_HOST` env variable.)

   ```diff
   - #set TOC(production,host) "10.10.10.10"  ;# IP of toc.oscar.aol.com
   + set TOC(production,host) "127.0.0.1"
   ```

7. **Sign on**

   Relaunch TiK and sign in!
