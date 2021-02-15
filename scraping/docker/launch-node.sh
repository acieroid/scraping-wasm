#!/bin/sh
if [ "$#" -ne 2 ]; then
    echo 'Requires the server URL as argument (e.g., 127.0.0.1:6345 or 10.0.0.1:6345) AND the node URL as argument (e.g., 127.0.0.1:6346 or 10.0.0.1:6346)'
    exit 1
fi

echo 'Building node'
./build-node.sh || exit 1
echo 'Launching it'
SERVER_URL="$1"
NODE_URL="$2"
PORT="$(echo $NODE_URL | cut -d: -f2)"
docker run -p "$PORT:$PORT" --dns 8.8.8.8 --dns 8.8.4.4 -it node ./node "$SERVER_URL" "$NODE_URL"
