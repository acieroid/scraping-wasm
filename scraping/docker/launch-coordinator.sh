#!/bin/sh
if [ "$#" -ne 1 ]; then
    echo 'Requires the server URL as argument (e.g., 127.0.0.1:6345 or 10.0.0.1:6345)'
    exit 1
fi

echo 'Building coordinator'
./build-coordinator.sh || exit 1
echo 'Launching it'
SERVER_URL="$1"
PORT="$(echo $SERVER_URL | cut -d: -f2)"
docker run -p "$PORT:$PORT" -d -v coordinatorout:/tmp/out coordinator ./coordinator "$SERVER_URL"

