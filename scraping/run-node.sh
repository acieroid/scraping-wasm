#!/bin/sh
if [ ! -f bin/node ]; then
    echo 'node not compiled, run make.sh'
    exit 1
fi

if [ "$#" -ne 2 ]; then
    if [ "$#" -ne 1 ]; then
        echo 'Running the node with the default URLs'
        SERVER_URL='127.0.0.1:6345'
    else
        echo 'Running the node with the given server URL but the default node URL'
        SERVER_URL="$1"
    fi
    NODE_URL='127.0.0.1:6346'
else
    SERVER_URL="$1"
    NODE_URL="$2"
fi

./bin/node $SERVER_URL $NODE_URL
    
