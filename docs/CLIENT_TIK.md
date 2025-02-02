# TiK Client Setup

This guide explains how to install and configure **TiK** for Retro AIM Server.

[TiK](https://en.wikipedia.org/wiki/TiK) is an open source instant messenger developed by AOL in the late 1990's. It
communicates using TOC, a text-based alternative to the OSCAR protocol.

In order to run Tik, you'll need to install the [Tcl/Tk runtime](https://www.tcl-lang.org/).

Installation guides are available for the following operating systems:

* [Windows](#windows)
* [Linux](#linux)
* [macOS (Intel & Apple Silicon)](#macos-intel--apple-silicon)

## Installation

### Windows

### Linux

### macOS (Intel & Apple Silicon)

1. **Download TiK**

   Download TiK from [Sourceforge](https://sourceforge.net/projects/tik/files/tik/) and extract the archive. The
   following are the most notable releases:

    - [v0.75](https://sourceforge.net/projects/tik/files/tik/0.75/) is the last official version developed by AOL.
    - [v0.90](https://sourceforge.net/projects/tik/files/tik/0.90/) is the last community release (recommended).

2. **Install Tcl/Tk 8.x**

   Open a new terminal and install Tcl/Tk 8.x using [Homebrew](https://brew.sh/).

    ```shell
   brew install tcl-tk@8
   ```

3. **Verify Tcl/Tk Version**

   macOS comes by default with a version of Tcl/Tk that does not support TiK. Verify that version 8.x installed in
   the previous step is in your PATH. Run the following command in a new terminal:

   ```shell
   echo 'puts $tcl_version' | tclsh
   ```

   If you not see version 8.x, try running `brew link tcl-tk@8` or `brew doctor` to fix the installation.

4. **Launch Tik**

   Open a terminal, navigate to the extracted archive, and launch TiK using `wish`.

   ```shell
   wish ./tik.tcl
   ```

   If all goes well, you should be greeted by the sign on window.

    <p align="center">
       <img width="400" alt="screenshot of TiK sign on window" src="https://github.com/user-attachments/assets/55b2d662-7ee8-4ec9-a49f-976799293bd7">
    </p>

5. **Configure Hostname**

   Configure TiK to connect to Retro AIM Server. From the sign on screen, click the `Configure` button. In the
   `TOC Host` field, enter the value of `TOC_HOST` from your server's `settings.env` config and click `OK`.

    <p align="center">
       <img width="400" alt="screenshot of TiK sign on window" src="https://github.com/user-attachments/assets/f2cdeede-2c0c-4c93-8dcc-335e6405ed6a">
    </p>

6. **Sign on**

   At this point, TiK should be able to successfully sign on.

