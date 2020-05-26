package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"distributed-key-value-store/config"
	cpb "distributed-key-value-store/protos/config"
	flpb "distributed-key-value-store/protos/follower"
)

var (
	testSpecPath = flag.String("test_spec", "", "The Test Spec to run")
	connTimeout  = flag.Int("timeout", 5, "Timeout, in seconds, when connecting to the follower")
)

// sentRequest
func sendRequest(ctx context.Context, c flpb.FollowerClient, t time.Time) {
	// rand.Seed(time.Now().UnixNano())
	// req := &flpb.PutRequest{
	// 	Key:     testKeys[rand.Intn(len(testKeys))],
	// 	Value:   uuid.New().String(),
	// 	Version: "1",
	// }
	// log.Printf("sending request %v", req)
	// resp, err := c.Put(ctx, req)
	// if err != nil {
	// 	log.Fatalf("%v.Put(_) = _, %v: ", c, err)
	// }
	// log.Println(resp)
}

func sendPutRequest(client flpb.FollowerClient, followerID string, putStep cpb.PutStep) {
	ctx := context.Background()
	req := &flpb.PutRequest{
		Key:   putStep.GetKey(),
		Value: putStep.GetValue(),
	}
	_, err := client.Put(ctx, req)
	if err != nil {
		log.Fatalf("%v.Put(_) = _, %v: ", followerID, err)
	}
}

func sendGetRequest(client flpb.FollowerClient, followerID string, gsetStep cpb.GetStep) {
	ctx := context.Background()
	req := &flpb.GetRequest{
		Key: gsetStep.GetKey(),
	}
	resp, err := client.Get(ctx, req)
	if err != nil {
		log.Fatalf("%v.Get(_) = _, %v: ", followerID, err)
	}

	if resp.Values[0] != gsetStep.GetAssertedValue() {
		log.Fatalf("%v.Get(_) %v != %v: ", resp.Values, gsetStep.GetAssertedValue())
	}

}

// sentRequests puts new key-value pair
func sentRequests(client flpb.FollowerClient) {
	// log.Printf("Sending request per 2 seconds")
	// ctx := context.Background()
	// ticker := time.NewTicker(2 * time.Second)
	// defer ticker.Stop()
	// done := time.After(time.Duration(*testTime) * time.Second)
	// for {
	// 	select {
	// 	case <-done:
	// 		log.Println("done sending requests")
	// 		return
	// 	case t := <-ticker.C:
	// 		sendRequest(ctx, client, t)
	// 	}
	// }
}

func readTestSpec() *cpb.TestSpec {
	content, err := ioutil.ReadFile(*testSpecPath)
	if err != nil {
		log.Fatalf("Failed to read test spec file %v", err)
	}

	testSpec := &cpb.TestSpec{}
	if err := proto.UnmarshalText(string(content), testSpec); err != nil {
		log.Fatalf("Failed to parse test spec file %v", err)
	}

	return testSpec
}

func main() {
	flag.Parse()
	configuration := config.ReadConfiguration()
	testSpec := readTestSpec()

	connMap := make(map[string]flpb.FollowerClient)
	// Connect to the follower.
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(time.Duration(*connTimeout)*time.Second))
	for followerID, followerAddress := range configuration.FollowerAddresses {
		log.Printf("Start client connecting to follower %q at address %q", followerID, followerAddress)
		// Only supporting localhost for testing.
		conn, err := grpc.Dial(followerAddress, opts...)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		connMap[followerID] = flpb.NewFollowerClient(conn)
		defer conn.Close()
	}

	for _, step := range testSpec.GetTestSteps() {
		followerID := step.GetFollowerId()
		switch s := step.Step.(type) {
		case *cpb.TestStep_PutStep:
			log.Printf("%v.Put: %v\n", followerID, s)
			// sendPutRequest(connMap[followerID], followerID, *s.PutStep)
		case *cpb.TestStep_GetStep:
			log.Printf("%v.Get: %v\n", followerID, s)
			// sendGetRequest(connMap[followerID], followerID, *s.GetStep)
		}
	}
}
