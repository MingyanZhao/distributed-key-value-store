.PHONY: default
default: build

#
# Protobufs and GRPC
#

protos/leader/leader.pb.go: protos/leader/leader.proto
	protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos/leader

protos/follower/follower.pb.go: protos/follower/follower.proto
	protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos/follower

protos/config/config.pb.go: protos/config/config.proto
	protoc -I protos protos/config/config.proto --go_out=plugins=grpc:protos/config 

#
# Binaries. Anything using `go build` is FORCEd because `go build` takes care of
# Go dependencies and incremental building. We only need to include non-Go as
# dependncies of each binary target.
#

bin/leader: FORCE protos/config/config.pb.go protos/leader/leader.pb.go
	go build -o bin/leader leader/leader.go

bin/follower: FORCE protos/config/config.pb.go protos/leader/leader.pb.go protos/follower/follower.pb.go
	go build -o bin/follower follower/follower.go

bin/client: FORCE protos/config/config.pb.go protos/follower/follower.pb.go
	go build -o bin/client client/client.go

.PHONY: build
build: bin/leader bin/follower bin/client

.PHONY: clean
clean:
	-rm -rf bin
	-rm -f protos/*/*.pb.go

.PHONY: FORCE
FORCE:

#
# Docker rules
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
