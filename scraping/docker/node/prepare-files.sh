#!/bin/sh
if [[ ! -f ./node || (-f ../../node && ../../node -nt ./node) ]]; then
    if [ ! -f ../../node ]; then
        echo 'node must be built first with make.sh in ../../'
        exit 1
    else
        echo 'Importing node binary from ../../'
        cp ../../node ./
    fi
fi
