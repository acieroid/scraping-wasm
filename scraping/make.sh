#!/bin/sh
export GOPATH=$(pwd)
go get scraping
go get node
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' node
go get coordinator
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' coordinator
