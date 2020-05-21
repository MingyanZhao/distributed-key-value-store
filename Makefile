PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := $(shell which protoc)

default: build

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

leader.pb.go: protos/leader/leader.proto | $(PROTOC_GEN_GO) $(PROTOC)
	protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos/leader

follower.pb.go: protos/follower/follower.proto | $(PROTOC_GEN_GO) $(PROTOC)
	protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos/follower

leader: leader.pb.go
	go build -o bin/leader leader/leader.go

follower: follower.pb.go
	go build -o follower follower/follower.go

client: follower.pb.go
	go build -o client client/client.go

build: leader follower client

clean:
	rm -rf bin; rm -f protos/*/*.pb.go

