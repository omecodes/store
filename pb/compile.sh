#!/bin/bash

$PROTOCPATH/bin/protoc -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
  -I$GOPATH/src/github.com/google/protobuf \
  -I$GOPATH/src/github.com/golang/protobuf \
  -I$GOPATH/src \
  -I. \
  --go-grpc_out . --go-grpc_opt paths=source_relative \
  --go_out . --go_opt paths=source_relative\
  --grpc-gateway_out .\
  --grpc-gateway_opt logtostderr=true \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt generate_unbound_methods=true \
  *.proto