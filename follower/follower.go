package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"distributed-key-value-store/config"
	configpb "distributed-key-value-store/protos/config"
	pb "distributed-key-value-store/protos/follower"
	lpb "distributed-key-value-store/protos/leader"
)

var (
	followerID                              = flag.String("follower_id", "", "The Follower's Unique ID")
	configuration   *configpb.Configuration = nil
	followerAddress                         = ""
)

type follower struct {
	pb.UnimplementedFollowerServer

	followerID string
	store      map[string][]*data
	leader     lpb.LeaderClient
	conn       *grpc.ClientConn
	udpateChan chan *lpb.UpdateRequest
}

type data struct {
	value   string
	version int64
}

func newFollower(followerID string) (*follower, error) {
	f := &follower{
		followerID: followerID,
		store:      make(map[string][]*data),
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(configuration.LeaderAddress, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		return nil, err
	}
	f.conn = conn
	f.leader = lpb.NewLeaderClient(conn)
	return f, nil
}

func (f *follower) stop() {
	f.conn.Close()
}

func (f *follower) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	// Storing the data
	log.Printf("Received request %v", req)

	values := f.store[req.Key]
	var v int64 = 0
	if f.store[req.Key] != nil {
		v = (values[len(values)-1]).version
	}

	newData := &data{value: req.Value, version: v}
	f.store[req.Key] = append(f.store[req.Key], newData)
	res := fmt.Sprintf("key value pair added: %v -> %v", req.Key, f.store[req.Key])
	log.Println(res)

	// TODO: write the data to the local disk
	// Infomation that needs to be written:
	// key, value, state
	// The states are:
	// received, updated, synced

	// Synching with leader
	log.Println("Sync with global leader...")

	updateReq := &lpb.UpdateRequest{
		Key:        req.Key,
		Address:    followerAddress,
		Version:    newData.version,
		FollowerId: *proto.String(f.followerID),
	}
	udpateResp, err := f.leader.Update(ctx, updateReq)
	if err != nil {
		return nil, err
	}
	log.Printf("Follower %v received sync response: %v", f.followerID, udpateResp)

	// Reply to the clilent
	return &pb.PutResponse{
		Result: res,
	}, nil
}

func main() {
	flag.Parse()

	configuration = config.ReadConfiguration()
	followerAddress = configuration.FollowerAddresses[*proto.String(*followerID)]

	log.Printf("Starting follower server and listening on address %v", followerAddress)
	lis, err := net.Listen("tcp", followerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	f, err := newFollower(*followerID)
	if err != nil {
		log.Fatalf("failed to create new follower: %v", err)
	}

	pb.RegisterFollowerServer(grpcServer, f)
	grpcServer.Serve(lis)
}
