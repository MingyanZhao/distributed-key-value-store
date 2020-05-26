.PHONY: default
default: build

#
# Protobufs and GRPC
#

PROTO_OPTS := -I protos --go_out=plugins=grpc:protos --go_opt=paths=source_relative

protos/leader/leader.pb.go: protos/leader/leader.proto protos/config/config.proto
	protoc protos/leader/leader.proto $(PROTO_OPTS)

protos/follower/follower.pb.go: protos/follower/follower.proto
	protoc protos/follower/follower.proto $(PROTO_OPTS)

protos/config/config.pb.go: protos/config/config.proto
	protoc protos/config/config.proto $(PROTO_OPTS)

#
# Binaries. Anything using `go build` is FORCEd because `go build` takes care of
# Go dependencies and incremental building. We only need to include non-Go as
# dependncies of each binary target.
#

bin/leader: FORCE protos/config/config.pb.go protos/leader/leader.pb.go protos/follower/follower.pb.go
	go build -o bin/leader distributed-key-value-store/leader

bin/follower: FORCE protos/config/config.pb.go protos/leader/leader.pb.go protos/follower/follower.pb.go
	go build -o bin/follower distributed-key-value-store/follower

bin/client: FORCE protos/config/config.pb.go protos/follower/follower.pb.go
	go build -o bin/client distributed-key-value-store/client

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
