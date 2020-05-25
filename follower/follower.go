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
)

const followerIDFlag = "follower_id"
const timeoutFlag = "timeout"

// Flags.
var (
	followerID = flag.String(followerIDFlag, "", "The Follower's Unique ID")
	timeout    = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to the leader")
)

// follower implements the Follower service.
type follower struct {
	// TODO: Remove. This stubs the Follower methods we haven't implemented yet.
	// Probably should keep it. We get Unimplemented error for free.
	pb.UnimplementedFollowerServer

	// These are set during creation and are immutable.
	id      string
	address string
	done    chan bool // TODO: Unused.

	// TODO: What happens if we lose the connection to the leader?
	// Need to handle it. Probably need to take a look at https://github.com/grpc/grpc-go/tree/master/examples/features/keepalive
	leader lpb.LeaderClient
	conn   *grpc.ClientConn

	// TODO: store needs a mutex.
	store map[string]*data
}

type data struct {
	values []*value
	// Each key has a update channel, so that different keys are updated concurrently.
	// The values for the same key are not updated in parallel because of linearizable consideration.
	updateChan chan *lpb.UpdateRequest
	syncChan   chan *pb.SyncRequest
	done       chan bool
}

type value struct {
	val     string
	version int64
}

// Create a follower that is connected to the leader.
func newFollower(configuration *configpb.Configuration, followerID, followerAddress string) (follower, error) {
	// Connect to the leader
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(time.Duration(*timeout)*time.Second))
	conn, err := grpc.Dial(configuration.LeaderAddress, opts...)
	if err != nil {
		return follower{}, fmt.Errorf("fail to dial leader at address %q: %v", configuration.LeaderAddress, err)
	}
	return follower{
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
		f.store[req.Key] = &data{updateChan: make(chan *lpb.UpdateRequest), done: make(chan bool)}
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
		Key:        req.Key,
		Address:    f.address,
		Version:    newData.version,
		FollowerId: *proto.String(f.id),
	}
	f.store[req.Key].updateChan <- updateReq

	// Reply to the clilent
	return &pb.PutResponse{
		Result: res,
	}, nil
}

func (f *follower) handleSync(key string) error {
	log.Printf("waiting for sync request")
	// ctx := context.Background()
	syncChan := f.store[key].syncChan
	for {
		select {
		case sync := <-syncChan:
			log.Printf("needs a sync %v", sync)
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
	updateChan := f.store[key].updateChan
	syncChan := f.store[key].syncChan
	for {
		select {
		case <-done:
			log.Printf("key %v update channel closed", key)
			return nil
		case updateReq := <-updateChan:
			// there is an update request
			log.Printf("send update request %v", updateReq)
			udpateResp, err := f.leader.Update(ctx, updateReq)
			if err != nil {
				// TODO: handle the error properlly.
				// One update message failed but still need to continue.
				log.Printf("Follower %v failed %v", f.id, err)
			}

			globalVer := updateReq.Version
			values := f.store[key].values
			localVer := values[len(values)-1].version
			askForVers := askForVersions(globalVer, localVer)
			syncReq := &pb.SyncRequest{
				Key:    key,
				AskFor: askForVers,
				MyData: &pb.Mydata{},
			}
			syncChan <- syncReq
			log.Printf("Follower %v received udpate response: %v", f.id, udpateResp)
		}
	}
}

func main() {
	flag.Parse()

	// Get config and options.
	configuration := config.ReadConfiguration()
	followerAddress, ok := configuration.FollowerAddresses[*proto.String(*followerID)]
	if !ok {
		log.Fatalf("did not specify a follower ID with the %q flag", followerIDFlag)
	}

	// Create and start the follower service.
	lis, err := net.Listen("tcp", followerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	f, err := newFollower(configuration, *followerID, followerAddress)
	if err != nil {
		log.Fatalf("failed to create new follower: %v", err)
	}

	log.Printf("Starting follower server and listening on address %q", followerAddress)
	grpcServer := grpc.NewServer()
	pb.RegisterFollowerServer(grpcServer, &f)
	grpcServer.Serve(lis)
}
