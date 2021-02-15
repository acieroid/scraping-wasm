#!/bin/sh
if [ ! -f bin/coordinator ]; then
    echo 'coordinator not compiled, run make.sh'
    exit 1
fi

if [ ! -f urls.txt ]; then
    echo 'urls.txt does not exists'
    if [ ! -f ../urls.txt ]; then
        echo 'urls.txt does not exists in the parent directory neither, run the get_urls.sh script'
        exit 1
    else
        echo 'copying urls.txt from the parent directory'
        cp ../urls.txt ./
    fi
fi

if [ "$#" -ne 1 ]; then
    echo 'Running the coordinator with the default server URL'
    SERVER_URL='127.0.0.1:6345'
else
    SERVER_URL="$1"
fi

echo 'Shuffling URLs...'
shuf urls.txt -o urls.txt
echo "Launching coordinator with URL $SERVER_URL"
./bin/coordinator $SERVER_URL
