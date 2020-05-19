package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	pb "distributed-key-value-store/protos/follower"
	lpb "distributed-key-value-store/protos/leader"
)

var (
	// followerAddr = flag.String("follower_addr", "", "follower address in the format of host:port")
	followerPort = flag.Int("follower_port", 10000, "follower port")
	leaderAddr   = flag.String("leader_addr", "localhost:9999", "global leader address in the format of host:port")
)

type follower struct {
	pb.UnimplementedFollowerServer

	followerID string
	store      map[string]string
	leader     lpb.LeaderClient
}

func newFollower() *follower {
	s := &follower{
		followerID: uuid.New().String(),
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
	conn, err := grpc.Dial(*leaderAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	l := lpb.NewLeaderClient(conn)

	syncReq := &lpb.SyncRequest{
		Key:        req.Key,
		Address:    fmt.Sprintf("localhost:%d", *followerPort),
		FollowerID: f.followerID,
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

	log.Printf("Starting follower server and listening on port %v", *followerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *followerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFollowerServer(grpcServer, newFollower())
	grpcServer.Serve(lis)
}
