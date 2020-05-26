package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"distributed-key-value-store/config"
	cpb "distributed-key-value-store/protos/config"
	flpb "distributed-key-value-store/protos/follower"
	"distributed-key-value-store/util"
)

var (
	testSpecPath = flag.String("test_spec", "", "The Test Spec to run")
	connTimeout  = flag.Int("timeout", 5, "Timeout, in seconds, when connecting to the follower")
)

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

func sendGetRequest(client flpb.FollowerClient, followerID string, getStep cpb.GetStep) {
	ctx := context.Background()
	req := &flpb.GetRequest{
		Key: getStep.GetKey(),
	}
	resp, err := client.Get(ctx, req)
	if err != nil {
		log.Fatalf("%v.Get(_) = _, %v: ", followerID, err)
	}

	if !reflect.DeepEqual(resp.GetValues(), getStep.GetAssertedValues()) {
		log.Fatalf("%v.Get(_) %v != %v: ", followerID, resp.Values, getStep.GetAssertedValues())
	}

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
	for followerID, followerAddress := range configuration.Followers {
		log.Printf("Start client connecting to follower %q at address %q", followerID, followerAddress)
		// Only supporting localhost for testing.
		conn, err := grpc.Dial(util.FormatServiceAddress(followerAddress), opts...)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		connMap[followerID] = flpb.NewFollowerClient(conn)
		defer conn.Close()
	}

	for _, step := range testSpec.GetTestSteps() {
		followerID := step.GetFollowerId()
		if step.GetPreSleepMs() > 0 {
			log.Printf("Sleeping for %v ms", step.GetPreSleepMs())
			time.Sleep(time.Duration(step.GetPreSleepMs()) * time.Millisecond)
		}
		switch s := step.Step.(type) {
		case *cpb.TestStep_PutStep:
			log.Printf("%v.Put: %v\n", followerID, s)
			sendPutRequest(connMap[followerID], followerID, *s.PutStep)
		case *cpb.TestStep_GetStep:
			log.Printf("%v.Get: %v\n", followerID, s)
			sendGetRequest(connMap[followerID], followerID, *s.GetStep)
		}
	}
}
