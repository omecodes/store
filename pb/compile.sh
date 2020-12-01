#!/bin/bash

$PROTOCPATH/bin/protoc --go_opt paths=source_relative --go_out . pb.proto