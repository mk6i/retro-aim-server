#!/bin/sh
# This script launches Retro AIM Server using go run with the environment vars
# defined in config/settings.env under MacOS/Linux. The script can be run from
# any working directory--it assumes the location of config/command files
# relative to the path of this script.
set -e

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ENV_FILE="$SCRIPT_DIR/../config/settings.env"
REPO_ROOT="$SCRIPT_DIR/.."

# Run Retro AIM Server from repo root.
cd "$REPO_ROOT"
go run -v ./cmd/server -config "$ENV_FILE"