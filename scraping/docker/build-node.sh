#!/bin/sh
cd node
./prepare-files.sh || exit 1

docker build -t node .
