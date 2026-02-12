#!/bin/bash
set -e

# Start VNC server in background
/usr/local/bin/start-vnc.sh &

# Execute nanobot with all passed arguments
exec /usr/local/bin/nanobot "$@"
