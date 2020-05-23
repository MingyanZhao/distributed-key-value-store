.PHONY: default
default: build

# Just rebuild the protos each time, it's basically instant and keeps the rule
# simple.
.PHONY: protos
protos:
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

.PHONY: build
build: bin/leader bin/follower bin/client

.PHONY: clean
clean:
	-rm -rf bin
	-rm -f protos/*/*.pb.go

#
# Docker rules.
#

.PHONY: docker
docker: docker-leader docker-follower docker-client

.PHONY: docker-leader
docker-leader: docker-base
	docker build -f Dockerfile_leader -t dkvs-leader .

.PHONY: docker-follower
docker-follower: docker-base
	docker build -f Dockerfile_follower -t dkvs-follower .

.PHONY: docker-client
docker-client: docker-base
	docker build -f Dockerfile_client -t dkvs-client .

.PHONY: docker-base
docker-base:
	docker build -f Dockerfile_base -t dkvs-base .
