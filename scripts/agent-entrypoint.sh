#!/bin/bash
set -euo pipefail

NANOBOT_HOME=/home/nanobot
NANOBOT_PATH="${NANOBOT_HOME}/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
DISPLAY_NUM=99
VNC_PORT=5900
WEBSOCKET_PORT=6080
INITIAL_RESOLUTION="${BROWSER_VIEW_INITIAL_RESOLUTION:-1600x2000}"
MAX_RESOLUTION="${BROWSER_VIEW_MAX_RESOLUTION:-4096x4096}"
DEPTH=24

should_start_browser_stack() {
	for arg in "$@"; do
		if [ "${arg}" = "--disable-ui" ]; then
			return 1
		fi
	done
	return 0
}

prepare_runtime_dirs() {
	mkdir -p /data "${NANOBOT_HOME}/.nanobot/state" "${NANOBOT_HOME}/sessions"
	mkdir -p /tmp/.X11-unix
	chmod 1777 /tmp/.X11-unix
	chown -R nanobot:nanobot /data "${NANOBOT_HOME}/.nanobot" "${NANOBOT_HOME}/sessions"
}

configure_fluxbox() {
	local fluxbox_dir="${HOME}/.nanobot/fluxbox"
	mkdir -p "${fluxbox_dir}"

	cat > "${fluxbox_dir}/init" <<'EOF'
session.screen0.toolbar.visible: false
session.screen0.tabs.intitlebar: false
session.screen0.tab.placement: TopLeft
session.screen0.fullMaximization: true
session.screen0.maxDisableMove: true
session.screen0.maxDisableResize: true
EOF

	cat > "${fluxbox_dir}/apps" <<'EOF'
[startup] {xsetroot -solid "#111111"}
[app] (name=.*)
  [Deco] {NONE}
  [Maximized] {yes}
  [Layer] {Normal}
  [FocusHidden] {yes}
  [IconHidden] {yes}
[end]
EOF
}

start_browser_stack() {
	local app_port="${NANOBOT_RUN_LISTEN_ADDRESS##*:}"
	if [ -z "${app_port}" ] || [ "${app_port}" = "${NANOBOT_RUN_LISTEN_ADDRESS}" ]; then
		app_port=8080
	fi

	export DISPLAY=":${DISPLAY_NUM}"

	echo "Starting Xvfb on :${DISPLAY_NUM}..."
	Xvfb ":${DISPLAY_NUM}" -screen 0 "${MAX_RESOLUTION}x${DEPTH}" +extension RANDR -ac -nolisten tcp &

	sleep 2

	configure_fluxbox

	echo "Starting window manager (fluxbox)..."
	fluxbox -rc "${HOME}/.nanobot/fluxbox/init" &

	echo "Starting x11vnc on port ${VNC_PORT}..."
	x11vnc -display ":${DISPLAY_NUM}" \
		-forever \
		-shared \
		-xrandr resize \
		-rfbport "${VNC_PORT}" \
		-nopw \
		-noxdamage \
		-no6 \
		-bg \
		-o /tmp/x11vnc.log

	echo "Starting websockify on port ${WEBSOCKET_PORT}..."
	websockify "${WEBSOCKET_PORT}" "localhost:${VNC_PORT}" &

	echo "Setting initial browser display size to ${INITIAL_RESOLUTION}..."
	if ! xrandr --display ":${DISPLAY_NUM}" --fb "${INITIAL_RESOLUTION}" >/dev/null 2>&1; then
		echo "WARNING: initial browser display resize to ${INITIAL_RESOLUTION} failed; continuing with Xvfb default size"
	fi

	echo "VNC server started!"
	echo "  - Display: :${DISPLAY_NUM}"
	echo "  - VNC port: ${VNC_PORT}"
	echo "  - Internal WebSocket port: ${WEBSOCKET_PORT}"
	echo "  - Access via Nanobot UI BrowserView on http://localhost:${app_port}"
}

if [ "$(id -u)" = "0" ]; then
	prepare_runtime_dirs
	exec runuser -u nanobot -- env \
		HOME="${NANOBOT_HOME}" \
		PATH="${NANOBOT_PATH}" \
		/usr/local/bin/nanobot "$@"
fi

export HOME="${NANOBOT_HOME}"
export PATH="${NANOBOT_PATH}"

mkdir -p "${NANOBOT_HOME}/.nanobot/state" "${NANOBOT_HOME}/sessions"

if should_start_browser_stack "$@"; then
	start_browser_stack
fi

exec /usr/local/libexec/nanobot "$@"
