package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	config "distributed-key-value-store/config"
	cpb "distributed-key-value-store/protos/config"
	pb "distributed-key-value-store/protos/leader"
)

// Flags.
const (
	casterFlag          = "caster"
	casterFlagImmediate = "immediate"
	casterFlagPeriodic  = "periodic"

	timeoutFlag = "timeout"
)

var (
	caster  = flag.String(casterFlag, casterFlagImmediate, fmt.Sprintf("Which broadcaster to use: %q or %q", casterFlagImmediate, casterFlagPeriodic))
	timeout = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to each follower")
)

type keyvaluemap struct {
	// TODO: per key automic update?
	m    sync.RWMutex
	data map[string]*versioninfo
}

// leader implements the leader service.
type leader struct {
	// TODO: Remove. This stubs the Follower methods we haven't implemented yet.
	// Probably should keep it. We get Unimplemented error for free.
	pb.UnimplementedLeaderServer
	configuration *cpb.Configuration
	keyVersionMap keyvaluemap

	// broadcaster sends updates to all followers when an update occurs.
	broadcaster broadcaster
}

type versioninfo struct {
	// version is the newest version the leader has updated.
	version int64

	// followerAddr is the address of the current primary follower.
	followerAddr string
}

func newLeader(configuration *cpb.Configuration) (*leader, error) {
	// Pick a broadcast mechanism.
	var bc broadcaster
	var err error
	switch *caster {
	case casterFlagImmediate:
		if bc, err = newImmediate(configuration); err != nil {
			return nil, err
		}
	case casterFlagPeriodic:
		if bc, err = newPeriodic(configuration); err != nil {
			return nil, err
		}
	default:
		panic(fmt.Sprintf("unknown caster: %q", *caster))
	}

	return &leader{
		configuration: configuration,
		keyVersionMap: keyvaluemap{data: make(map[string]*versioninfo)},
		broadcaster:   bc,
	}, nil
}

func (l *leader) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	log.Printf("leader received update request %v", req)
	l.keyVersionMap.m.Lock()
	defer l.keyVersionMap.m.Unlock()

	// Update the version info for the key.
	if l.keyVersionMap.data[req.Key] == nil {
		l.keyVersionMap.data[req.Key] = &versioninfo{version: 0}
	}
	result := pb.UpdateResult_SUCCESS
	if l.keyVersionMap.data[req.Key].version > req.Version {
		result = pb.UpdateResult_NEED_SYNC
	}

	l.keyVersionMap.data[req.Key].version++
	l.keyVersionMap.data[req.Key].followerAddr = req.Address

	// Send a SUCCESS response with the new version and followers.
	// TODO: add support for NEED_SYNC case
	var primaryFollowerID, backupFollowerID string
	primaryFollowerID = req.FollowerId
	backupFollowerID = req.FollowerId
	resp := &pb.UpdateResponse{
		Version: l.keyVersionMap.data[req.Key].version,
		Result:  result,
		PrePrimary: &pb.FollowerEndpoint{
			Address:    l.configuration.FollowerAddresses[primaryFollowerID],
			FollowerId: primaryFollowerID,
		},
		PreBackup: &pb.FollowerEndpoint{
			Address:    l.configuration.FollowerAddresses[backupFollowerID],
			FollowerId: backupFollowerID,
		},
	}

	// There's been a successful update. Notify followers.
	l.broadcaster.enqueue(req.Key, resp)

	log.Printf("leader replying update response %v", resp)
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
	ldr, err := newLeader(configuration)
	if err != nil {
		log.Fatalf("failed to create leader: %v", err)
	}
	pb.RegisterLeaderServer(grpcServer, ldr)
	grpcServer.Serve(lis)
}
