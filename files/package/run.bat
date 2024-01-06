@echo off
REM This script launches Retro AIM Server. By default, it assumes that the
REM executable and configuration file are located in the same directory as this
REM script.

setlocal enabledelayedexpansion

REM Get the directory of the script
for %%I in (%0) do set "SCRIPT_DIR=%%~dpI"
set "ENV_FILE=!SCRIPT_DIR!\settings.env"
set "EXEC_FILE=!SCRIPT_DIR!\retro-aim-server.exe"

REM Load the settings file
if exist !ENV_FILE! (
    call "!ENV_FILE!"
) else (
    echo error: environment file '!ENV_FILE!' not found.
    exit /b 1
)

REM Start Retro AIM Server
if exist !EXEC_FILE! (
    call "!EXEC_FILE!"
) else (
    echo error: executable '!EXEC_FILE!' not found.
    exit /b 1
)
