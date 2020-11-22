#!/bin/bash

$PROTOCPATH/bin/protoc --go_out . --go_opt paths=source_relative store.proto