#!/usr/bin/env bash

# Generate the protos.
protoc -I/usr/local/include -I. \
       -I$GOPATH/src \
       --go_out=plugins=grpc:./go/. \
       rpc.proto

protoc -I/usr/local/include -I. \
        --js_out=import_style=commonjs,binary:./js-node/. \
        --grpc_out=./js-node/. \
        --plugin=protoc-gen-grpc=$(which grpc_tools_node_protoc_plugin) \
        rpc.proto


# Dependencies:
# npm install ts-protoc-gen

protoc -I/usr/local/include -I. \
       --js_out=import_style=commonjs,binary:./js-web/. \
       --js_service_out=./js-web/. \
       --plugin=protoc-gen-js_service=$(which protoc-gen-js_service) \
       rpc.proto

protoc  -I/usr/local/include -I. \
        --js_out=import_style=commonjs,binary:./ts-web/. \
        --ts_out=service=true:./ts-web/. \
        --plugin=protoc-gen-ts=$(which protoc-gen-ts) \
        rpc.proto

### Generate the REST reverse prozxy.
#protoc -I/usr/local/include -I. \
#        -I$GOPATH/src \
#       -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#       --grpc-gateway_out=logtostderr=true:./go/. \
#       rpc.proto
#
## Finally, generate the swagger file which describes the REST API in detail.
#protoc -I/usr/local/include -I. \
#       -I$GOPATH/src \
#       -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#       --swagger_out=logtostderr=true:. \
#       rpc.proto