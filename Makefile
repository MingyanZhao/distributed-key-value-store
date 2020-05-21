PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := $(shell which protoc)

default: build

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

protos: $(wildcard protos/*) $(PROTOC_GEN_GO) $(PROTOC)
	protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos/leader
	protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos/follower
	protoc -I protos protos/config/config.proto --go_out=plugins=grpc:protos/config 

LEADER=$(wildcard leader/*)
FOLLOWER=$(wildcard follower/*)
CLIENT=$(wildcard client/*)
CONFIG=$(wildcard config/*)

bin/leader:  $(LEADER) protos
	go build -o bin/leader leader/leader.go

bin/follower: $(FOLLOWER) $(CONFIG) protos
	go build -o bin/follower follower/follower.go

bin/client: $(CLIENT) $(CONFIG) protos
	go build -o bin/client client/client.go

build: bin/leader bin/follower bin/client

clean:
	rm -rf bin; rm -f protos/*/*.pb.go

