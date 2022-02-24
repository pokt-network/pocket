#!/bin/sh

config=$1

# TODO(olshansky): Make sure this works on everyone's workstations with Docker & Goland debugging.

if [ -z "$DEBUG_PORT" ]; then
    echo "DEBUG DISABLED"
    command="go run cmd/pocket/main.go --config=$config"
else
    echo "DEBUG ENABLED on port $DEBUG_PORT"
    command="touch /tmp/output.dlv && dlv debug cmd/pocket/main.go --headless --accept-multiclient --listen=:$DEBUG_PORT --api-version=2 --continue --output /tmp/output.dlv --  --config=$config"
fi

reflex \
  --start-service \
  -r '\.go' \
  -s -- sh -c "$command";
