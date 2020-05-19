package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	flpb "distributed-key-value-store/protos/follower"
)

var (
	followerAddr = flag.String("follwer_addr", "localhost:10000", "The server address in the format of host:port")
	followerPort = flag.Int("follwer_port", 10000, "The follower port")

	testKeys = []string{"key-0", "Key-1", "Key-2", "Key-3", "key-4", "key-5", "key-6", "key-7", "key-8", "key-9"}
)

// sentRequest
func sendRequest(ctx context.Context, c flpb.FollowerClient, t time.Time) {
	rand.Seed(time.Now().UnixNano())
	req := &flpb.AppendRequest{
		Key:     testKeys[rand.Intn(len(testKeys))],
		Value:   uuid.New().String(),
		Version: "1",
	}
	log.Printf("sending request %v", req)
	resp, err := c.Append(ctx, req)
	if err != nil {
		log.Fatalf("%v.Append(_) = _, %v: ", c, err)
	}
	log.Println(resp)
}

// sentRequests puts new key-value pair
func sentRequests(client flpb.FollowerClient) {
	log.Printf("Sending request per 2 seconds")
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	ctx := context.Background()
	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				log.Printf("client closed")
				return
			case t := <-ticker.C:
				sendRequest(ctx, client, t)
			}
		}
	}()
	time.Sleep(2 * time.Minute)
	ticker.Stop()
	done <- true
	log.Println("stop sending request")
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	log.Println("Start client...")
	// only supporting localhost
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", *followerPort), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := flpb.NewFollowerClient(conn)

	// Send requests
	sentRequests(client)
}
