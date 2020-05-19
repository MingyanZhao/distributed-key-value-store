# distributed-key-value-store
A highly available distributed key value store.

# Prerequisites
1. Clone the repository under `$GOPATH/src`

1. Install dependencies (may need more)
```bash
go get -u github.com/golang/protobuf/protoc-gen-go
go get github.com/google/uuid
```

1. Generate proto files
```
protoc -I protos protos/follower/follower.proto --go_out=plugins=grpc:protos/follower
protoc -I protos protos/leader/leader.proto --go_out=plugins=grpc:protos/leader
```

## Start components

1. The leader waits for the sync signal from the followers and sends back sync responses.
Start the leader:
```bash
go run leader/leader.go
```

2. Follower listens to port 10000 for incoming client requests and periodically syncs up with the leader.
Start the follower:
```bash
go run follower/follower.go
```
Optionally, set a different port
```bash
go run client/client.go --follwer_port=10001
```

3. Start a test client sending request to a follower.
Start the test client with default port, 10000
```bash
go run client/client.go
```
Optionally, set a different port
```bash
go run client/client.go --follwer_port=10001
```
