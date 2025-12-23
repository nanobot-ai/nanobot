#!/bin/bash
# Entrypoint script for sandbox container
# Takes a JSON configuration as the first argument

set -e -x

CONFIG_JSON="$1"

if [ -z "$CONFIG_JSON" ]; then
	echo "Error: No configuration provided" >&2
	exit 1
fi

# Parse JSON using jq
BASE_URI=$(echo "$CONFIG_JSON" | jq -r '.baseUri // empty')
WORKDIR=$(echo "$CONFIG_JSON" | jq -r '.workdir // "/workspace"')

# Always use "sandbox" as the ID
SANDBOX_ID="sandbox"

echo "=== Sandbox Initialization ==="
echo "ID: $SANDBOX_ID"
echo "Base URI: $BASE_URI"
echo "Workdir: $WORKDIR"
echo "=============================="

# Check if database exists at /.data/.agentfs/sandbox.db
DB_FILE="/.data/.agentfs/${SANDBOX_ID}.db"

if [ ! -f "$DB_FILE" ]; then
	echo "No database found, initializing new agentfs database..."
	cd /.data
	BASE_ARGS=""
	if [ -d /base ]; then
	  echo "Using /base as base directory for initialization"
	  BASE_ARGS="--base /base"
  fi
	if ! agentfs init "$SANDBOX_ID" --force $BASE_ARGS 2>&1; then
		echo "Error: Failed to initialize database" >&2
		exit 1
	fi
	echo "Database initialized: $DB_FILE"
else
	echo "Database already exists, skipping initialization"
fi

if [ ! -f "$DB_FILE" ]; then
	echo "Error: No database file found after initialization" >&2
	ls -la /.data/.agentfs/ || echo "/.data/.agentfs/ does not exist"
	exit 1
fi

echo "Using database: $DB_FILE"

# Handle baseUri if it's a git URL
if [ -n "$BASE_URI" ]; then
	if [[ "$BASE_URI" =~ ^(https?://|git@|git://) ]]; then
		echo "Detected git URL in baseUri, cloning repository..."
		# Extract repo name from URL for directory name
		REPO_NAME=$(basename "$BASE_URI" .git)
		CLONE_DIR="/.data/$REPO_NAME"

		if [ ! -d "$CLONE_DIR" ]; then
			git clone "$BASE_URI" "$CLONE_DIR"
			echo "Repository cloned to $CLONE_DIR"
		else
			echo "Repository already exists at $CLONE_DIR"
		fi
	else
		echo "baseUri is not a git URL, skipping clone"
	fi
fi


if [ -d /base ]; then
  sleep infinity
  exit 1
fi

# Create mount point if it doesn't exist
mkdir -p "$WORKDIR"


# Mount agentfs to the workspace and run in foreground
echo "Mounting agentfs at $WORKDIR..."
exec agentfs mount \
	--foreground \
	--auto-unmount \
	--allow-root \
	--uid 10000 \
	--gid 10000 \
	"$DB_FILE" \
	"$WORKDIR"
