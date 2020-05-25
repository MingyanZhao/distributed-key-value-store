package main

import (
	"context"
	"flag"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	config "distributed-key-value-store/config"
	cpb "distributed-key-value-store/protos/config"
	pb "distributed-key-value-store/protos/leader"
)

type keyvaluemap struct {
	// TODO: per key automic update?
	m    sync.RWMutex
	data map[string]*versioninfo
}

type leader struct {
	pb.UnimplementedLeaderServer
	configuration *cpb.Configuration
	keyVersionMap keyvaluemap
}

type versioninfo struct {
	version      int64
	followerAddr string
}

func newLeader(configuration *cpb.Configuration) *leader {
	l := &leader{
		configuration: configuration,
		keyVersionMap: keyvaluemap{data: make(map[string]*versioninfo)},
	}
	return l
}

func (l *leader) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	log.Printf("leader received update request %v", req)
	l.keyVersionMap.m.Lock()
	defer l.keyVersionMap.m.Unlock()
	if l.keyVersionMap.data[req.Key] == nil {
		l.keyVersionMap.data[req.Key] = &versioninfo{version: 0}
	}
	l.keyVersionMap.data[req.Key].version++
	l.keyVersionMap.data[req.Key].followerAddr = req.Address

	var primaryFollowerID, backupFollowerID string
	primaryFollowerID = req.FollowerId
	// TODO: update backup follower selection procedure.
	backupFollowerID = req.FollowerId
	resp := &pb.UpdateResponse{
		Version: l.keyVersionMap.data[req.Key].version,
		Result:  pb.UpdateResult_SUCCESS,
		PrePrimary: &pb.FollowerEndpoint{
			Address:    l.configuration.FollowerAddresses[primaryFollowerID],
			FollowerId: primaryFollowerID,
		},
		PreBackup: &pb.FollowerEndpoint{
			Address:    l.configuration.FollowerAddresses[backupFollowerID],
			FollowerId: backupFollowerID,
		},
	}
	log.Printf("leader replying udpate response %v", resp)
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
	pb.RegisterLeaderServer(grpcServer, newLeader(configuration))
	grpcServer.Serve(lis)
}
