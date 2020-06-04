package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	"distributed-key-value-store/config"
	flpb "distributed-key-value-store/protos/follower"
	"distributed-key-value-store/util"
)

const followerIDFlag = "follower_id"
const timeoutFlag = "timeout"
const modeFlag = "mode"

var (
	followerID = flag.String(followerIDFlag, "", "The Follower to talk to")
	mode       = flag.String(modeFlag, "normal", "client mode")
	timeout    = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to the follower")
	testTime   = flag.Int("testtime", 30, "How long to send updates for, in seconds")

	threadCount  = flag.Int("threadcount", 1, "The number of threads that sends a number of requests")
	requestCount = flag.Int("requestcount", 10, "The number of requests to send")
	testKeys     = []string{"key-0", "key-1", "key-2", "key-3", "key-4", "key-5", "key-6", "key-7", "key-8", "key-9"}
	// testKeys   = []string{"key-0"}
	valueCount = 0
	countLock  = sync.Mutex{}
)

// sentRequest
func sendRequest(ctx context.Context, c flpb.FollowerClient, t time.Time) {
	rand.Seed(time.Now().UnixNano())
	k := testKeys[rand.Intn(len(testKeys))]
	countLock.Lock()
	req := &flpb.PutRequest{
		Key:   k,
		Value: fmt.Sprintf("client-%v-%v-value-%d", *followerID, k, valueCount),
	}
	valueCount++
	countLock.Unlock()

	log.Printf("sending request %v", req)
	resp, err := c.Put(ctx, req)
	if err != nil {
		log.Fatalf("%v.Put(_) = _, %v: ", c, err)
	}
	log.Println(resp)
}

// sentRequest
func sendConcurRequest(ctx context.Context, c flpb.FollowerClient, t time.Time) {
	rand.Seed(time.Now().UnixNano())
	k := testKeys[rand.Intn(len(testKeys))]
	req := &flpb.PutRequest{
		Key:   k,
		Value: fmt.Sprintf("client-%v-%v-value-%s", *followerID, k, uuid.New()),
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
	log.Printf("Sending request per 1 seconds")
	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Second)
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

func getData(c flpb.FollowerClient) string {
	ctx := context.Background()
	// var result = make([][]string, 10)
	log.Printf("********************************")
	var result string
	for _, k := range testKeys {
		req := &flpb.GetRequest{
			Key: k,
		}
		resp, err := c.Get(ctx, req)
		if err != nil {
			log.Fatalf("%v.Get(_) = _, %v: ", c, err)
		}
		sort.Strings(resp.Values)
		// result[i] = resp.Values
		r := fmt.Sprintf("%v: [%v]\n\n\n", k, resp.Values)
		result += r
	}
	log.Printf("********************************")
	return result
}

func runClient(client flpb.FollowerClient) {
	// Send requests
	sentRequests(client)

	// Get the data
	log.Printf("Get data first time")
	getData(client)

	// Wait for other test clients to finish
	time.Sleep(10 * time.Second)

	// Get the data
	log.Printf("Get data second time")
	getData(client)
}

func sendRequestsConcurrently(ctx context.Context, c flpb.FollowerClient) {
	for i := 0; i < *requestCount; i++ {
		go sendConcurRequest(ctx, c, time.Now())
	}
}

func perfTest(c flpb.FollowerClient) {
	ctx := context.Background()
	for i := 0; i < *threadCount; i++ {
		go sendRequestsConcurrently(ctx, c)
	}

	// Wait for other test clients to finish
	time.Sleep(10 * time.Second)

	log.Printf("Get data second time")
	result := getData(c)
	fileName := fmt.Sprintf("./perfresult-follower-%s", *followerID)
	err := ioutil.WriteFile(fileName, []byte(result), 0644)
	if err != nil {
		panic(err)
	}
}

func main() {
	log.Printf("Start client connecting to follower - 1")
	flag.Parse()
	log.Printf("Start client connecting to follower - 2")
	configuration := config.ReadConfiguration()
	log.Printf("Start client connecting to follower - 3")

	// There must be a follower specified via flag, and that follower must be in
	// the config.
	followerAddress, ok := configuration.Followers[*proto.String(*followerID)]
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
	conn, err := grpc.Dial(util.FormatServiceAddress(followerAddress), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := flpb.NewFollowerClient(conn)
	log.Printf("client mode is %v", *mode)
	switch *mode {
	case "normal":
		runClient(client)
	case "perf":
		perfTest(client)
	}

}
