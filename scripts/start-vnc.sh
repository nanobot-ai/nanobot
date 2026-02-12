#!/bin/bash
# Start VNC server for browser-use remote viewing

set -e

# Configuration
DISPLAY_NUM=99
VNC_PORT=5900
WEBSOCKET_PORT=6080
RESOLUTION="1920x1080"
DEPTH=24

echo "Starting Xvfb on :${DISPLAY_NUM}..."
Xvfb :${DISPLAY_NUM} -screen 0 ${RESOLUTION}x${DEPTH} -ac -nolisten tcp &
XVFB_PID=$!

# Wait for Xvfb to be ready
sleep 2

# Set DISPLAY for subsequent commands
export DISPLAY=:${DISPLAY_NUM}

echo "Starting window manager (fluxbox)..."
fluxbox &
FLUXBOX_PID=$!

echo "Starting x11vnc on port ${VNC_PORT}..."
x11vnc -display :${DISPLAY_NUM} \
    -forever \
    -shared \
    -rfbport ${VNC_PORT} \
    -nopw \
    -noxdamage \
    -no6 \
    -bg \
    -o /tmp/x11vnc.log

echo "Starting websockify on port ${WEBSOCKET_PORT}..."
websockify --web=/usr/share/novnc ${WEBSOCKET_PORT} localhost:${VNC_PORT} &
WEBSOCKIFY_PID=$!

echo "VNC server started!"
echo "  - Display: :${DISPLAY_NUM}"
echo "  - VNC port: ${VNC_PORT}"
echo "  - WebSocket port: ${WEBSOCKET_PORT}"
echo "  - Access via: http://localhost:${WEBSOCKET_PORT}/vnc.html"

# Keep script running
wait
