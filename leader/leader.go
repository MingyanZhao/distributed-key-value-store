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
	"distributed-key-value-store/util"
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
	// Different keys should be updated concurrently. So each key should have a lock.
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

	fInfo *followerInfo
}

type followerInfo struct {
	// followerAddr is the address of the follower.
	followerAddr *cpb.ServiceAddress
	followerID   string
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
		log.Printf("leader first seen this key, create a new entry, %v", req.Key)
		l.keyVersionMap.data[req.Key] = &versioninfo{
			version: 0,
			fInfo: &followerInfo{
				followerAddr: l.configuration.Followers["0"],
				followerID:   "0",
			},
		}
	} else {
		log.Printf("leader has the key %v, version %v, follower info %v", req.Key, l.keyVersionMap.data[req.Key].version, *l.keyVersionMap.data[req.Key].fInfo)
	}

	// Set the result
	result := pb.UpdateResult_SUCCESS
	if l.keyVersionMap.data[req.Key].version > req.Version {
		result = pb.UpdateResult_NEED_SYNC
	}

	// Advance the version number for the key
	l.keyVersionMap.data[req.Key].version++

	// Update the follower info. Swtich the current primary to the input follower info.
	prePrimary := *l.keyVersionMap.data[req.Key].fInfo
	log.Printf("preprimary is %v, id %v", prePrimary.followerAddr.Port, prePrimary.followerID)

	l.keyVersionMap.data[req.Key].fInfo.followerAddr = req.FollowerAddress
	l.keyVersionMap.data[req.Key].fInfo.followerID = req.FollowerId
	log.Printf("new preprimary is %v id %v", l.keyVersionMap.data[req.Key].fInfo.followerAddr.Port, l.keyVersionMap.data[req.Key].fInfo.followerID)
	log.Printf("preprimary again is %v, id %v", prePrimary.followerAddr.Port, prePrimary.followerID)

	resp := &pb.UpdateResponse{
		Version: l.keyVersionMap.data[req.Key].version,
		Result:  result,
		PrePrimary: &pb.FollowerEndpoint{
			Address:    prePrimary.followerAddr,
			FollowerId: prePrimary.followerID,
		},
		// TODO: implement backup follower setting. Now the primary and the backup are the same
		PreBackup: &pb.FollowerEndpoint{
			Address:    prePrimary.followerAddr,
			FollowerId: prePrimary.followerID,
		},
	}

	// There's been a successful update. Notify followers.
	// Not sure if we want to notify every follower in the update message.
	// We definitely can! But Braodcast here means the followers would do a lot of sync on every request.
	// Comment it out temporarily for testing.
	// l.broadcaster.enqueue(req.Key, resp)

	log.Printf("leader replying update response %v", resp)
	return resp, nil
}

func main() {
	flag.Parse()

	configuration := config.ReadConfiguration()

	log.Printf("Starting leader server and listening on address %v", configuration.Leader)
	lis, err := net.Listen("tcp", util.FormatServiceAddress(configuration.Leader))
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
