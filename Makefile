PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := $(shell which protoc)

default: build

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

protos/leader/leader.pb.go: protos/leader/leader.proto | $(PROTOC_GEN_GO) $(PROTOC)
	protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos/leader

protos/follower/follower.pb.go: protos/follower/follower.proto | $(PROTOC_GEN_GO) $(PROTOC)
	protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos/follower

leader: protos/leader/leader.pb.go
	go build -o bin/leader leader/leader.go

follower: protos/leader/leader.pb.go protos/follower/follower.pb.go
	go build -o bin/follower follower/follower.go

client: protos/follower/follower.pb.go
	go build -o bin/client client/client.go

build: leader follower client

clean:
	rm -rf bin; rm -f protos/*/*.pb.go

