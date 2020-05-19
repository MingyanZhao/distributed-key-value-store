package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "distributed-key-value-store/protos/leader"
)

var (
	leaderAddr = flag.String("leader_addr", "localhost:9999", "The server address in the format of host:port")
	leaderPort = flag.Int("leader_port", 9999, "The server port")
)

type leader struct {
	pb.UnimplementedLeaderServer

	keyVersionMap map[string]*versioninfo
}

type versioninfo struct {
	version      int64
	followerAddr string
}

func newLeader() *leader {
	l := &leader{
		keyVersionMap: make(map[string]*versioninfo),
	}
	return l
}

func (l *leader) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	log.Printf("leader received sync request %v", req)
	if l.keyVersionMap[req.Key] == nil {
		l.keyVersionMap[req.Key] = &versioninfo{version: 0}
	}
	l.keyVersionMap[req.Key].version++
	l.keyVersionMap[req.Key].followerAddr = req.Address

	resp := &pb.SyncResponse{
		Version: l.keyVersionMap[req.Key].version,
		Result:  "TBD",
		TargetFollowers: []*pb.TargetFollower{
			{
				Address: req.Address,
			},
		},
	}
	log.Printf("leader replying sync response %v", resp)
	return resp, nil
}

func main() {
	flag.Parse()

	log.Printf("Starting leader server and listening on port %v", *leaderPort)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *leaderPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterLeaderServer(grpcServer, newLeader())
	grpcServer.Serve(lis)
}
