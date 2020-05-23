package main

import (
	"context"
	"flag"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	config "distributed-key-value-store/config"
	pb "distributed-key-value-store/protos/leader"
)

type keyvaluemap struct {
	// TODO: per key automic update?
	m    sync.RWMutex
	data map[string]*versioninfo
}

type leader struct {
	pb.UnimplementedLeaderServer

	keyVersionMap keyvaluemap
}

type versioninfo struct {
	version      int64
	followerAddr string
}

func newLeader() *leader {
	l := &leader{
		keyVersionMap: keyvaluemap{data: make(map[string]*versioninfo)},
	}
	return l
}

func (l *leader) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	log.Printf("leader received sync request %v", req)
	l.keyVersionMap.m.Lock()
	defer l.keyVersionMap.m.Unlock()
	if l.keyVersionMap.data[req.Key] == nil {
		l.keyVersionMap.data[req.Key] = &versioninfo{version: 0}
	}
	l.keyVersionMap.data[req.Key].version++
	l.keyVersionMap.data[req.Key].followerAddr = req.Address

	resp := &pb.SyncResponse{
		Version: l.keyVersionMap.data[req.Key].version,
		Result:  "TBD",
		TargetFollowers: []*pb.TargetFollower{
			{
				Address:    req.Address,
				FollowerId: req.FollowerId,
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
