#!/bin/sh
# This script launches Retro AIM Server under MacOS/Linux. Because it assumes
# that the executable and settings.env file are located in the same directory
# as this script, the script can be run from any directory.
set -e

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ENV_FILE="$SCRIPT_DIR/settings.env"
EXEC_FILE="$SCRIPT_DIR/retro_aim_server"

# Load the settings file.
if [ -f "$ENV_FILE" ]; then
    . "$ENV_FILE"
else
    echo "error: environment file '$ENV_FILE' not found."
    exit 1
fi

# Start Retro AIM Server.
if [ -f "$EXEC_FILE" ]; then
    "$EXEC_FILE"
else
    echo "error: executable '$EXEC_FILE' not found."
    exit 1
fi