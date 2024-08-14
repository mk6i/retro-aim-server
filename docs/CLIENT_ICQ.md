# ICQ Client Setup

This guide explains how to install and configure **ICQ 2000b** for Retro AIM Server.

> ICQ 2000b is the only version of ICQ currently supported by RAS.

## Installation

### Windows

1. **Download ICQ**

   Download ICQ 2000b from [archive.org](https://archive.org/details/icq2000b_202206).

2. **Run Installer**

   Install ICQ as you would normally install any application. For Windows 10/11,
   set [Windows XP compatibility mode](https://support.microsoft.com/en-us/windows/make-older-apps-or-programs-compatible-with-windows-783d6dd7-b439-bdb0-0490-54eea0f45938)
   on the executable.

3. **Close the Registration Window**

   Once installation is complete, you'll be greeted by an ICQ registration window.
   *Do not complete the registration wizard.* Close the window and move on to
   the [post-installation steps](#post-install-configuration).

    <p align="center">
       <img width="400" alt="screenshot of ICQ registration window" src="https://github.com/user-attachments/assets/b5684b93-02b0-4314-adfa-16ea9826cf69">
    </p>

### Linux

You can ICQ 2000b under Linux via [WINE](https://www.winehq.org/).

1. **Download ICQ**

   Download ICQ 2000b from [archive.org](https://archive.org/details/icq2000b_202206).

2. **Install WINE**

   Run and install [WINE](https://wiki.winehq.org/Download)

3. **Run the Installer**

   Start the ICQ installer under WINE from a terminal:
   ```shell
   wine icq2000b.exe
   ```

4. **Close the Registration Window**

   Once installation is complete, you'll be greeted by an ICQ registration window.
   *Do not complete the registration wizard.* Close the window and move on to
   the [post-installation steps](#post-install-configuration).

    <p align="center">
       <img width="400" alt="screenshot of ICQ registration window" src="https://github.com/user-attachments/assets/b5684b93-02b0-4314-adfa-16ea9826cf69">
    </p>

### macOS (Intel & Apple Silicon)

Windows ICQ 2000b can run on modern macOS (including the Apple Silicon platform) without a VM
using [WineskinServer](https://github.com/Gcenx/WineskinServer), a wrapper for WINE.

1. **Install WineskinServer**

   Install WineskinServer via homebrew:

   ```shell
   brew install --cask --no-quarantine gcenx/wine/wineskin
   ```

2. **Create a Blank Application Wrapper**

   Launch `Wineskin Winery`. Install the latest engine and create a new blank
   wrapper for installing ICQ.

   Generating the wrapper might take 1-2 minutes, and the application might not
3. respond during this time. Once complete, click `View wrapper in Finder`.

   <p align="center">
      <img width="325" alt="screenshot of wineskin server" src="https://github.com/user-attachments/assets/8d2ed477-f41c-4f00-90c2-d24c468b4aae">
   </p>

3. **Install ICQ into the application wrapper**

   Launch the wrapper from the Finder window. Select `Install Software`.

   <p align="center">
      <img width="325" alt="screenshot of wineskin server" src="https://github.com/user-attachments/assets/a2c0882c-230f-4034-9670-6f309ec1f628">
   </p>

   Select `Choose Setup Executable` and open the ICQ installer executable.

   <p align="center">
      <img width="325" alt="screenshot of wineskin server" src="https://github.com/user-attachments/assets/cf6aa78b-6bc0-4cfe-b750-bf2c763acdfb">
   </p>

4. **Run the Installer**

   Complete the ICQ installation wizard.

5. **Close the Registration Window**

   Once installation is complete, you'll be greeted by an ICQ registration window.
   *Do not complete the registration wizard.* Close the window and move on to
   the [post-installation steps](#post-install-configuration).

    <p align="center">
       <img width="400" alt="screenshot of ICQ registration window" src="https://github.com/user-attachments/assets/b5684b93-02b0-4314-adfa-16ea9826cf69">
    </p>

## Post-install Configuration

Configure the default OSCAR hostname in the registry so that ICQ can connect to
your server. This needs to be done before the first launch of the ICQ.

> Do not attempt to set the ICQ hostname via the registration Window. If you do
> this, a bug will surface that forever prevents the client from "remembering"
>  settings such as saved passwords and OSCAR hostname.  

1. **Open Registry Editor**

    - Windows
        - Click `Start > Run`.
        - Enter `regedit` and click `OK`.
    - Wine (Linux)
        - Open a terminal.
        - Run `wine regedit` in a terminal.
    - WineskinServer (macOS)
        - Open a terminal.
        - Run the following command, substituting `icq2000b.app` with the file name of your wrapper:
          ```shell
          ~/Applications/Wineskin/icq2000b.app/Contents/Wineskin.app/Contents/Resources/regedit
          ```
      
2. **Open Default ICQ Settings**

   Navigate to `HKEY_CURRENT_USER\Software\Mirabilis\ICQ\DefaultPrefs`.
   <p align="center">
      <img width="500" alt="screenshot of regedit" src="https://github.com/user-attachments/assets/02b20e3a-769c-4c69-bbf5-395684d8f30f">
   </p>

3. **Configure OSCAR Host**

   Set the `Default Server Host` registry entry to the value of `OSCAR_HOST`
   found in `config/settings`.
   <p align="center">
      <img width="325" alt="screenshot editing Default Server Host in regedit" src="https://github.com/user-attachments/assets/ebcf66fa-1841-41f7-986a-90b24dd0a94d">
   </p>

4. **Configure Server Port**

   Set `Default Server Port` to the value of `AUTH_PORT` found in `config/settings`. Make sure to tick the `Decimal`
   radio button.
   <p align="center">
      <img width="325" alt="screenshot editing Default Server Port in regedit" src="https://github.com/user-attachments/assets/11a3efff-40f1-4f1d-b88a-9c78fddb9c3d">
   </p>
   
5. **Exit Registry Editor**

   Client configuration is complete. Close the Registry Editor.   

## First Time Login

Now that ICQ is installed and configured, all that remains is to log in.

Start ICQ and complete the first-time registration wizard. Start by selecting `Existing User`.

> Do not try to create a new user in the registration wizard. To create a new user in Retro AIM Server, follow account creation steps in the [server quickstart guides](https://github.com/mk6i/retro-aim-server?tab=readme-ov-file#-how-to-run).
   
   <p align="center">
      <img width="400" alt="screenshot of ICQ registration wizard" src="https://github.com/user-attachments/assets/93bd3fc9-96a6-45ff-af90-e96b3f938dc3">
   </p>


Enter the ICQ user credentials and finish out the rest of the wizard.

<p align="center">
   <img width="400" alt="screenshot of ICQ registration wizard" src="https://github.com/user-attachments/assets/7520db7c-0512-42d1-88f3-e3f8f9d5eaec">
</p>

You should now be able to successfully connect ICQ 2000b to Retro AIM Server.