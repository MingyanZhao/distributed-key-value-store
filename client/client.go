package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	"distributed-key-value-store/config"
	flpb "distributed-key-value-store/protos/follower"
)

const followerIDFlag = "follower_id"
const timeoutFlag = "timeout"

var (
	followerID = flag.String(followerIDFlag, "", "The Follower to talk to")
	timeout    = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to the follower")
	testTime   = flag.Int("testtime", 120, "How long to send updates for, in seconds")

	testKeys = []string{"key-0", "key-1", "key-2", "key-3", "key-4", "key-5", "key-6", "key-7", "key-8", "key-9"}
)

// sentRequest
func sendRequest(ctx context.Context, c flpb.FollowerClient, t time.Time) {
	rand.Seed(time.Now().UnixNano())
	req := &flpb.PutRequest{
		Key:     testKeys[rand.Intn(len(testKeys))],
		Value:   uuid.New().String(),
		Version: "1",
	}
	log.Printf("sending request %v", req)
	resp, err := c.Put(ctx, req)
	if err != nil {
		log.Fatalf("%v.Put(_) = _, %v: ", c, err)
	}
	log.Println(resp)
}

// sentRequests puts new key-value pair
func sentRequests(client flpb.FollowerClient) {
	log.Printf("Sending request per 2 seconds")
	ctx := context.Background()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	done := time.After(time.Duration(*testTime) * time.Second)
	for {
		select {
		case <-done:
			log.Println("done sending requests")
			return
		case t := <-ticker.C:
			sendRequest(ctx, client, t)
		}
	}
}

func main() {
	flag.Parse()
	configuration := config.ReadConfiguration()

	// There must be a follower specified via flag, and that follower must be in
	// the config.
	followerAddress, ok := configuration.FollowerAddresses[*proto.String(*followerID)]
	if !ok {
		log.Fatalf("did not specify a follower ID with the %q flag", followerIDFlag)
	}

	// Connect to the follower.
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(time.Duration(*timeout)*time.Second))
	log.Printf("Start client connecting to follower %q at address %q", *followerID, followerAddress)
	// Only supporting localhost for testing.
	conn, err := grpc.Dial(followerAddress, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := flpb.NewFollowerClient(conn)

	// Send requests
	sentRequests(client)
}
