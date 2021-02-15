#!/bin/sh
cd coordinator
./prepare-files.sh || exit 1

docker build -t coordinator .
docker volume create coordinatorout
