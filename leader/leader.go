package main

import (
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	config "distributed-key-value-store/config"
	cpb "distributed-key-value-store/protos/config"
	fpb "distributed-key-value-store/protos/follower"
	pb "distributed-key-value-store/protos/leader"
	"distributed-key-value-store/util"
)

// Flags.
const (
	casterFlag          = "caster"
	casterFlagImmediate = "immediate"
	casterFlagPeriodic  = "periodic"
	// Simple braodcasting means that the leader send a Notification of all of the key-version info to all of followers
	casterFlagSimple = "simple"
	timeoutFlag      = "timeout"

	keyVersionMapConcurrency = 32
)

var (
	caster  = flag.String(casterFlag, casterFlagImmediate, fmt.Sprintf("Which broadcaster to use: %q or %q", casterFlagImmediate, casterFlagPeriodic))
	timeout = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to each follower")
)

type keyvaluemap struct {
	// Improve concurrency by sharding the locks.
	// Avoid using sync.Map which is design for append-only and mostly-read cases
	// https://github.com/golang/go/issues/20360
	m    [keyVersionMapConcurrency]sync.RWMutex
	data map[string]*versioninfo
}

func (kvm *keyvaluemap) lockKey(key string) {
	slot := hashKey(key)
	kvm.m[slot].Lock()
}

func (kvm *keyvaluemap) unlockKey(key string) {
	slot := hashKey(key)
	kvm.m[slot].Unlock()
}

func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32() % keyVersionMapConcurrency
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

	done chan bool

	followers map[string]fpb.FollowerClient
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
	var follwers map[string]fpb.FollowerClient
	switch *caster {
	case casterFlagImmediate:
		if bc, err = newImmediate(configuration); err != nil {
			return nil, err
		}
	case casterFlagPeriodic:
		if bc, err = newPeriodic(configuration); err != nil {
			return nil, err
		}
	case casterFlagSimple:
		follwers, err = loadFollowers(configuration)
		if err != nil {
			return nil, err
		}
	default:
		panic(fmt.Sprintf("unknown caster: %q", *caster))
	}

	return &leader{
		configuration: configuration,
		keyVersionMap: keyvaluemap{data: make(map[string]*versioninfo)},
		broadcaster:   bc,
		followers:     follwers,
	}, nil
}

func (l *leader) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	log.Printf("leader received update request %v", req)
	l.keyVersionMap.lockKey(req.Key)
	defer l.keyVersionMap.unlockKey(req.Key)

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
		PrePrimary: &cpb.FollowerEndpoint{
			Address:    prePrimary.followerAddr,
			FollowerId: prePrimary.followerID,
		},
		// TODO: implement backup follower setting. Now the primary and the backup are the same
		PreBackup: &cpb.FollowerEndpoint{
			Address:    prePrimary.followerAddr,
			FollowerId: prePrimary.followerID,
		},
	}

	// There's been a successful update. Notify followers.
	// Not sure if we want to notify every follower in the update message.
	// Braodcasting here means the followers would do a lot of sync on every request.
	// Since the version will be commited after the follower recives the update response,
	// the key that is being asked for may not yet be commited and therefore broadcasting might not working as expected.
	// TODO: Comment it out temporarily for testing.
	// l.broadcaster.enqueue(req.Key, resp)

	log.Printf("leader replying update response %v", resp)
	return resp, nil
}

func loadFollowers(config *cpb.Configuration) (map[string]fpb.FollowerClient, error) {
	// Establish a GRPC client for each follower.
	followers := make(map[string]fpb.FollowerClient)
	for id, service := range config.Followers {
		conn, err := grpc.Dial(util.FormatServiceAddress(service), grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("failed to dial follower %q service %v: %v", id, service, err)
		}
		followers[id] = fpb.NewFollowerClient(conn)
	}
	return followers, nil
}

func (l *leader) broadcasting() {
	log.Printf("Sending broadcasting request per 2 seconds")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	ctx := context.Background()
	for {
		select {
		case <-l.done:
			log.Println("done broadcasting")
			return
		case <-ticker.C:
			log.Println("broadcast...")
			for id, fc := range l.followers {
				for k, vi := range l.keyVersionMap.data {
					if id == vi.fInfo.followerID {
						continue
					}
					req := &fpb.NotifyRequest{
						Key:     k,
						Version: vi.version,
						Primary: &cpb.FollowerEndpoint{
							Address:    vi.fInfo.followerAddr,
							FollowerId: vi.fInfo.followerID,
						},
						// TODO: Fix logic for backup follower
						Backup: &cpb.FollowerEndpoint{
							Address:    vi.fInfo.followerAddr,
							FollowerId: vi.fInfo.followerID},
					}
					fc.Notify(ctx, req)
				}
			}
		}
	}
}

func main() {
	flag.Parse()

	configuration := config.ReadConfiguration()

	log.Printf("Starting leader server and listening on address %v", configuration.Leader)
	lis, err := net.Listen("tcp", util.FormatBindAddress(configuration.Leader))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	ldr, err := newLeader(configuration)
	if err != nil {
		log.Fatalf("failed to create leader: %v", err)
	}

	if *caster == casterFlagSimple {
		go ldr.broadcasting()
	}

	pb.RegisterLeaderServer(grpcServer, ldr)
	grpcServer.Serve(lis)
}
