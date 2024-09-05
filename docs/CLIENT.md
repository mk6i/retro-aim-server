# AIM Client Setup

This guide explains how to install and configure AIM clients for Retro AIM Server.

## Installation

### Linux

Windows AIM versions 5.0-5.1.3036 run perfectly well on [Wine](https://www.winehq.org/). Here's how to set up AIM
5.1.3036.

1. [Install Wine](https://wiki.winehq.org/Download)
2. Download the AIM installer from [archive.org](https://archive.org/details/aim513036)
3. Run the installer from the terminal:
   ```shell
   wine aim324235.exe
   ```

### macOS (Intel & Apple Silicon)

Windows AIM can run on modern macOS without a VM, including the Apple Silicon platform!

Get started with the [AIM for macOS project](https://github.com/mk6i/aim-for-macos).

### Windows 10/11

All versions of AIM 5.x run on Windows 10/11 with varying degrees of success. To get started, install AIM
[5.0.2829](http://www.oldversion.com/windows/aol-instant-messenger-5-0-2829). Set [Windows XP compatibility mode](https://support.microsoft.com/en-us/windows/make-older-apps-or-programs-compatible-with-windows-783d6dd7-b439-bdb0-0490-54eea0f45938)
on the executable once installed.

Newer 5.x versions exhibit a quirk where `aim.exe` randomly hangs on startup, which can be mitigated by [AIM Tamer](http://iwarg.ddns.net/phoenix/index.php?action=downloads).

## Configuration

Once installed, configure AIM to connect to Retro AIM Server.

1. At the sign-on screen, click `Setup`.
   <p>
      <img width="319" alt="screenshot of AIM sign-on screen" src="https://github.com/mk6i/mkdb/assets/2894330/9e0e743e-e41d-4c45-9e82-d97d7d4325f3">
   </p>
2. Under the `Sign On/Off` category, click `Connection`.
   <p>
      <img width="662" alt="screenshot of AIM preferences window" src="https://github.com/mk6i/mkdb/assets/2894330/c7cfcaa4-8132-4b57-b5c9-7643c99cbda2">
   </p>
3. In the `Host` field, enter the value of `OSCAR_HOST` found in `config/settings`. In the `Port` field, enter the
   value of `AUTH_PORT` found in `config/settings.env`.
   <p>
      <img width="618" alt="Screen Shot 2024-03-29 at 11 22 19 PM" src="https://github.com/mk6i/mkdb/assets/2894330/da17c457-a773-4b82-b4ba-cb81f9a2e085">
   </p>
4. Click OK and sign in to AIM!