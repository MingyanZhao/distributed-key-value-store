# distributed-key-value-store
A highly available distributed key value store.

There are 3 processes:

* **leader** - waits for the sync signal from the followers and sends back sync
  responses.
* **follower** - listens to port 10000 for incoming client requests and
  periodically syncs up with the leader.
* **client** - a test client sending request to a follower. Start the test
  client with default port, 10000.

# Prerequisites

To run in docker, skip to [running with docker](#running-with-docker).

1. Install [the protobuf compiler](https://github.com/protocolbuffers/protobuf/releases).

2. Install [the Go protobuf plugin](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)

3. Make sure your path includes `$GOPATH/bin` (`$HOME/go/bin` by default).

4. The default configration is in `protos/config/local-config.pb.txt`; you can
   use your own config file to set a different port for example.

# Building, Testing and Running

## Using `make`

1. Run `make` to build the project.

2. Binaries are in bin/.

## Testing
Enter a foler, then run the test.
```bash
$ cd follower
$ go test
```

## Run manually 

1. Generate proto files
Run script:
```
./proto-gen.sh
```
Or generate them seprately:

```
protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos/follower
protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos/leader
protoc -I protos protos/config/config.proto --go_out=plugins=grpc:protos/config
```

2. You can start each process with `go run`:
```bash
go run distributed-key-value-store/leader
go run distributed-key-value-storeer/follower
go run distributed-key-value-store/client
```

# Running with Docker

1. Run `make docker` to build the three docker images: `dkvs-leader`,
   `dkvs-follower`, and `dkvs-client`.

2. Run them with docker, for example:
```bash
docker run --rm --network host dkvs-leader
```
# Running the tests

1. Run `test/run_test.sh` in the project dirctory