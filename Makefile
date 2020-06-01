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

protos/config/test_spec.pb.go: protos/config/test_spec.proto
	protoc -I protos protos/config/test_spec.proto --go_out=plugins=grpc:protos/config 

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

bin/test_client: FORCE protos/config/test_spec.pb.go protos/config/config.pb.go protos/follower/follower.pb.go
	go build -o bin/test_client test/test_client.go

.PHONY: build
build: bin/leader bin/follower bin/client bin/test_client

.PHONY: clean
clean:
	-rm -rf bin
	-rm -f protos/*/*.pb.go
	-rm -f paper/*.aux paper/*.log paper/*.pdf paper/*.bbl paper/*.blg

.PHONY: FORCE
FORCE:

#
# Paper
#

.PHONY: paper
paper: paper/paper.pdf

paper/paper.pdf: paper/paper.tex paper/paper.bib
	cd paper && \
		pdflatex paper && \
		bibtex paper && \
		pdflatex paper && \
		pdflatex --jobname=paper paper

.PHONY: spell
spell:
	aspell --mode=tex --home-dir=. --personal=aspell.en.pws --repl=aspell.en.prepl check paper/paper.tex

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
