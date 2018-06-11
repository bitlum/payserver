#!/usr/bin/env bash

# Generate the protos.
protoc -I/usr/local/include -I. \
       -I$GOPATH/src \
       --go_out=plugins=grpc:./go/. \
       rpc.proto
