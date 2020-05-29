#!/bin/bash

# Generate proto files
echo "Generating config proto"
protoc -I protos protos/config/config.proto --go_out=plugins=grpc:protos --go_opt=paths=source_relative
echo "Generating follower proto"
protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos --go_opt=paths=source_relative
echo "Generating leader proto"
protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos --go_opt=paths=source_relative