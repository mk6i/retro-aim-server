#!/bin/sh
# This script launches Retro AIM Server. By default, it assumes that the
# executable and configuration file are located in the same directory as this
# script.
set -e

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ENV_FILE="$SCRIPT_DIR/settings.env"
EXEC_FILE="$SCRIPT_DIR/retro-aim-server"

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