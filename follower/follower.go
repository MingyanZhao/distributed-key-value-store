package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"distributed-key-value-store/config"
	configpb "distributed-key-value-store/protos/config"
	pb "distributed-key-value-store/protos/follower"
	lpb "distributed-key-value-store/protos/leader"
	"distributed-key-value-store/util"
)

const followerIDFlag = "follower_id"
const timeoutFlag = "timeout"

// Flags.
var (
	followerID            = flag.String(followerIDFlag, "", "The Follower's Unique ID")
	timeout               = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to the leader")
	followerConnectionMap = make(map[string]pb.FollowerClient)
)

// follower implements the Follower service.
type follower struct {
	// TODO: Remove. This stubs the Follower methods we haven't implemented yet.
	// Probably should keep it. We get Unimplemented error for free.
	pb.UnimplementedFollowerServer

	// These are set during creation and are immutable.
	id      string
	address *configpb.ServiceAddress
	done    chan bool // TODO: Unused.

	// TODO: What happens if we lose the connection to the leader?
	// Need to handle it. Probably need to take a look at https://github.com/grpc/grpc-go/tree/master/examples/features/keepalive
	leader lpb.LeaderClient
	conn   *grpc.ClientConn

	// TODO: store needs a mutex.
	store map[string]*data
}

type data struct {
	// TODO: values could be a list https://golang.org/pkg/container/list/
	values []*value
	// Each key has a update channel, so that different keys are updated concurrently.
	// The values for the same key are not updated in parallel because of linearizable consideration.
	updateReqChan  chan *lpb.UpdateRequest
	updateRespChan chan *lpb.UpdateResponse
	done           chan bool
}

type value struct {
	val     string
	version int64
}

func dial(addr *configpb.ServiceAddress) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(time.Duration(*timeout)*time.Second))
	conn, err := grpc.Dial(util.FormatServiceAddress(addr), opts...)
	if err != nil {
		return nil, fmt.Errorf("fail to dial server at address %v: %v", addr, err)
	}
	return conn, nil
}

// Create a follower that is connected to the leader.
func newFollower(configuration *configpb.Configuration, followerID string, followerAddress *configpb.ServiceAddress) (*follower, error) {
	// Connect to the leader
	conn, err := dial(configuration.Leader)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to leader at address %v", configuration.Leader)
	return &follower{
		id:      followerID,
		address: followerAddress,
		store:   make(map[string]*data),
		done:    make(chan bool),
		conn:    conn,
		leader:  lpb.NewLeaderClient(conn),
	}, nil
}

// TODO: Unused.
func (f *follower) stop() {
	f.conn.Close()
}

func (f *follower) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	// Storing the data
	log.Printf("Received request %v", req)

	// Get the current version (or 0 if it's the first).
	var v int64 = 0
	if d, ok := f.store[req.Key]; ok {
		v = (d.values[len(d.values)-1]).version
	} else {
		// A new key is received. Start the channels for the new key.
		f.store[req.Key] = &data{
			updateReqChan:  make(chan *lpb.UpdateRequest),
			updateRespChan: make(chan *lpb.UpdateResponse),
			done:           make(chan bool),
		}
		go f.handleUpdate(req.Key)
		go f.handleSync(req.Key)
	}

	newData := &value{val: req.Value, version: v}
	f.store[req.Key].values = append(f.store[req.Key].values, newData)
	res := fmt.Sprintf("key value pair added: %v -> %v", req.Key, f.store[req.Key])
	log.Println(res)

	// TODO: write the data to the local disk
	// Infomation that needs to be written:
	// key, value, state
	// The states are:
	// received, updated, synced

	// TODO: Make this happen asynchronously.
	// Sending update with leader
	log.Println("Sync with global leader...")

	updateReq := &lpb.UpdateRequest{
		Key:             req.Key,
		FollowerAddress: f.address,
		Version:         newData.version,
		FollowerId:      *proto.String(f.id),
	}
	f.store[req.Key].updateReqChan <- updateReq

	// Reply to the clilent
	return &pb.PutResponse{
		Result: res,
	}, nil
}

// TODO: DUMMY IMPL PLEASE FIX!
func (f *follower) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	var values []string
	if d, ok := f.store[req.Key]; ok {
		values = make([]string, len(f.store[req.Key].values))
		for i, v := range d.values {
			values[i] = v.val
		}
	} else {
		values = []string{}
	}

	return &pb.GetResponse{
		Values: values,
	}, nil
}

// If the values is a list https://golang.org/pkg/container/list/, we could avoid this.
func (f *follower) latestVersion(key string) int64 {
	values := f.store[key].values
	return values[len(values)-1].version
}

func (f *follower) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	log.Printf("Received Sync request %v", req)

	return &pb.SyncResponse{
		Key: req.Key,
		// TODO: the values filed needs some refactoring.
		// TODO: implement SyncResponse logic
		Value: []*pb.Value{
			{
				Value: []string{"test-1", "test-2"},
				// Version: f.latestVersion(req.Key),
				Version: 78,
			},
		},
	}, nil
}

func (f *follower) handleSync(key string) error {
	log.Printf("waiting for sync request")
	// ctx := context.Background()
	updateRespChan := f.store[key].updateRespChan
	ctx := context.Background()
	for {
		select {
		case updateResp := <-updateRespChan:
			// Send sync request to the target follower
			log.Printf("needs a sync %v", updateResp)
			values := f.store[key].values
			localVer := values[len(values)-1].version
			globalVer := updateResp.Version
			askForVers := askForVersions(globalVer, localVer)

			syncReq := &pb.SyncRequest{
				Key:    key,
				AskFor: askForVers,
				MyData: &pb.Mydata{},
			}

			if _, ok := followerConnectionMap[updateResp.PrePrimary.FollowerId]; !ok {
				newConn, err := dial(updateResp.PrePrimary.Address)
				if err != nil {
					// TODO: handle error
					log.Printf("error: failed to connect the pre-primary %v, at address: %v", updateResp.PrePrimary.FollowerId, updateResp.PrePrimary.Address)
					continue
				}
				followerConnectionMap[updateResp.PrePrimary.FollowerId] = pb.NewFollowerClient(newConn)
			}
			prePrimary := followerConnectionMap[updateResp.PrePrimary.FollowerId]

			syncResp, err := prePrimary.Sync(ctx, syncReq) // Send the sync to the pre-primary followerConnectionMap
			if err != nil {
				// TODO: handle error
				log.Printf("Sync rpc failed, %v", err)
				continue
			}
			log.Printf("SyncResp received, %v", syncResp)
		}
	}
}

func askForVersions(globalVer, localVer int64) *pb.AskFor {
	result := &pb.AskFor{}
	var i int64
	for i = localVer + 1; i < globalVer; i++ {
		result.Versions = append(result.Versions, i)
	}
	return result
}

func (f *follower) handleUpdate(key string) error {
	log.Printf("waiting for update request")
	ctx := context.Background()
	done := f.store[key].done
	updateReqChan := f.store[key].updateReqChan
	updateRespChan := f.store[key].updateRespChan
	for {
		select {
		case <-done:
			log.Printf("key %v update channel closed", key)
			return nil
		case updateReq := <-updateReqChan:
			// there is an update request
			log.Printf("send update request %v", updateReq)
			updateResp, err := f.leader.Update(ctx, updateReq)
			if err != nil {
				// TODO: handle the error properlly.
				// One update message failed but still need to continue.
				log.Printf("Follower %v failed %v", f.id, err)
			}
			log.Printf("Follower %v received update response: %v", f.id, updateResp)
			if updateResp.Result == lpb.UpdateResult_NEED_SYNC {
				updateRespChan <- updateResp
			}
		}
	}
}

// TODO: Handle.
func (f *follower) Notify(ctx context.Context, req *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	log.Printf("Got a notification")
	return &pb.NotifyResponse{
		Success: true,
	}, nil
}

func main() {
	flag.Parse()

	// Get config and options.
	configuration := config.ReadConfiguration()
	followerAddress, ok := configuration.Followers[*proto.String(*followerID)]
	if !ok {
		log.Fatalf("did not specify a follower ID with the %q flag", followerIDFlag)
	}

	// Create and start the follower service.
	lis, err := net.Listen("tcp", util.FormatServiceAddress(followerAddress))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	f, err := newFollower(configuration, *followerID, followerAddress)
	if err != nil {
		log.Fatalf("failed to create new follower: %v", err)
	}

	log.Printf("Starting follower server and listening on address %v", followerAddress)
	grpcServer := grpc.NewServer()
	pb.RegisterFollowerServer(grpcServer, f)
	grpcServer.Serve(lis)
}
