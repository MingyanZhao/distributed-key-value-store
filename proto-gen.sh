#!/bin/bash

# Generate proto files
echo "Generating config proto"
protoc -I protos protos/config/config.proto --go_out=plugins=grpc:protos/config
echo "Generating follower proto"
protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos/follower
echo "Generating leader proto"
protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos/leader