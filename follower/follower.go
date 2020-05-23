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
	followerID                              = flag.Int("follower_id", 0, "The Follower's Unique ID")
	configuration   *configpb.Configuration = nil
	followerAddress                         = ""
)

type follower struct {
	pb.UnimplementedFollowerServer

	followerID int
	store      map[string]string
	leader     lpb.LeaderClient
}

func newFollower(followerID int) *follower {
	s := &follower{
		followerID: followerID,
		store:      make(map[string]string),
	}
	return s
}

func (f *follower) Append(ctx context.Context, req *pb.AppendRequest) (*pb.AppendResponse, error) {
	// Storing the data
	log.Printf("Received request %v", req)
	f.store[req.Key] = req.Value
	res := fmt.Sprintf("key value pair added: %v -> %v", req.Key, f.store[req.Key])
	log.Println(res)

	// Synching with leader
	log.Println("Sync with global leader...")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(configuration.LeaderAddress, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	l := lpb.NewLeaderClient(conn)

	syncReq := &lpb.SyncRequest{
		Key:        req.Key,
		Address:    followerAddress,
		FollowerId: *proto.Int(f.followerID),
	}
	syncResp, err := l.Sync(ctx, syncReq)
	if err != nil {
		return nil, err
	}
	log.Printf("Follower %v received sync response: %v", f.followerID, syncResp)

	// Reply to the clilent
	return &pb.AppendResponse{
		Result: res,
	}, nil
}

func main() {
	flag.Parse()

	configuration = config.ReadConfiguration()
	followerAddress = configuration.FollowerAddresses[*proto.Int(*followerID)]

	log.Printf("Starting follower server and listening on address %v", followerAddress)
	lis, err := net.Listen("tcp", followerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFollowerServer(grpcServer, newFollower(*followerID))
	grpcServer.Serve(lis)
}
