package main

import (
	"context"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"

	config "distributed-key-value-store/config"
	pb "distributed-key-value-store/protos/leader"
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

	configuration := config.ReadConfiguration()

	log.Printf("Starting leader server and listening on address %v", configuration.LeaderAddress)
	lis, err := net.Listen("tcp", configuration.LeaderAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterLeaderServer(grpcServer, newLeader())
	grpcServer.Serve(lis)
}
