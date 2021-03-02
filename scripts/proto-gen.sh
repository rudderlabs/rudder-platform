#!/bin/bash

scriptdir=$( cd "${0%/*}" && pwd)
binpath=$( cd "$scriptdir"/.. && pwd )/bin

gen() {
  for var in "$@"
  do
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative "$var"
  done
}

go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0 \
         google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0

gen proto/rudder-sources/*.proto \
    proto/common/*.proto \
    proto/warehouse/*.proto

