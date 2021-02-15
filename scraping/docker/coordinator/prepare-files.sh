#!/bin/sh
if [[ ! -f ./coordinator || (-f ../../coordinator && ../../coordinator -nt ./coordinator) ]]; then
    if [ ! -f ../../coordinator ]; then
        echo 'coordinator must be built first with make.sh in ../../'
        exit 1
    else
        echo 'Importing coordinator binary from ../../'
        cp ../../coordinator ./
    fi
fi

if [ ! -f urls.txt ]; then
    if [ ! -f ../../urls.txt ]; then
       echo 'Cannot find urls.txt file either here nor in ../../'
       exit 1
    else
        echo 'Importing URLs'
        cp ../../urls.txt ./
    fi
fi
