#!/bin/sh

IMPORT_PATH=$GOPATH/src/github.com/gogo/protobuf/protobuf:$GOPATH/src:.
protoc --proto_path=$IMPORT_PATH --gofast_out . *proto