package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
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
const logOutputFlag = "logOutput"

var (
	followerID = flag.String(followerIDFlag, "", "The Follower to talk to")
	mode       = flag.String(modeFlag, "normal", "client mode")
	timeout    = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to the follower")
	testTime   = flag.Int("testtime", 30, "How long to send updates for, in seconds")

	threadCount  = flag.Int("threadcount", 2, "The number of threads that sends a number of requests")
	requestCount = flag.Int("requestcount", 10, "The number of requests to send")
	testKeys     = []string{"key-0", "key-1", "key-2", "key-3", "key-4", "key-5", "key-6", "key-7", "key-8", "key-9"}
	// testKeys   = []string{"key-0"}
	valueCount = 0
	countLock  = sync.Mutex{}

	elapsedLock = sync.Mutex{}
	sendElapsed = int64(0)

	logOutput = flag.String(logOutputFlag, "log-only", "Path to the log file")
	logger    *log.Logger
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

	logger.Printf("sending request %v", req)
	resp, err := c.Put(ctx, req)
	if err != nil {
		logger.Fatalf("%v.Put(_) = _, %v: ", c, err)
	}
	logger.Println(resp)
}

// sentRequest
func sendConcurRequest(ctx context.Context, c flpb.FollowerClient, t time.Time, thread int) {
	rand.Seed(time.Now().UnixNano())
	k := testKeys[rand.Intn(len(testKeys))]
	req := &flpb.PutRequest{
		Key:   k,
		Value: fmt.Sprintf("client-%v-%v-t-%d-%s", *followerID, k, thread, uuid.New()),
	}

	logger.Printf("sending request %v", req)
	start := time.Now()
	_, err := c.Put(ctx, req)
	elapsed := time.Since(start)
	elapsedLock.Lock()
	valueCount++
	sendElapsed += elapsed.Microseconds()
	elapsedLock.Unlock()
	if err != nil {
		logger.Fatalf("%v.Put(_) = _, %v: ", c, err)
	}
	// logger.Printf("client(follower id) %v, got response %v", *followerID, resp)
}

// sentRequests puts new key-value pair
func sentRequests(client flpb.FollowerClient) {
	logger.Printf("Sending request per 1 seconds")
	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	done := time.After(time.Duration(*testTime) * time.Second)
	for {
		select {
		case <-done:
			logger.Println("done sending requests")
			return
		case t := <-ticker.C:
			sendRequest(ctx, client, t)
		}
	}
}

func getData(c flpb.FollowerClient) string {
	ctx := context.Background()
	// var result = make([][]string, 10)
	logger.Printf("********************************")
	var received int
	var result string
	start := time.Now()

	for _, k := range testKeys {
		req := &flpb.GetRequest{
			Key: k,
		}

		resp, err := c.Get(ctx, req)
		if err != nil {
			logger.Fatalf("%v.Get(_) = _, %v: ", c, err)
		}
		received += len(resp.Values)
		sort.Strings(resp.Values)
		r := fmt.Sprintf("%v: [%v]\n\n\n", k, resp.Values)
		result += r
	}
	elapsed := time.Since(start)
	perRequest := float64(elapsed.Microseconds()) / float64(len(testKeys))
	logger.Printf("get %d values in total, elapsed %v, per request %v microseconds", received, elapsed, perRequest)
	logger.Printf("********************************")
	return result
}

func runClient(client flpb.FollowerClient) {
	// Send requests
	sentRequests(client)

	// Get the data
	logger.Printf("Get data first time")
	getData(client)

	// Wait for other test clients to finish
	time.Sleep(10 * time.Second)

	// Get the data
	logger.Printf("Get data second time")
	getData(client)
}

func sendRequestsConcurrently(ctx context.Context, c flpb.FollowerClient, t int) {
	logger.Printf("this is routine %d", t)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	done := make(chan bool)
	count := *requestCount
	for {
		select {
		case <-done:
			logger.Println("done perf test sending requests")
			return
		case <-ticker.C:
			go sendConcurRequest(ctx, c, time.Now(), t)
			count--
			if count == 0 {
				done <- true
			}
		}
	}
}

func perfTest(c flpb.FollowerClient) {
	ctx := context.Background()
	for i := 0; i < *threadCount; i++ {
		go sendRequestsConcurrently(ctx, c, i)
	}

	// Wait for other test clients to finish
	time.Sleep(time.Duration(*testTime) * time.Second)

	perRequest := float64(sendElapsed) / float64(*requestCount)
	logger.Printf("********************************")
	logger.Printf("sent %d values in total, elapsed %v us, per request %v us", (*requestCount)*(*threadCount), sendElapsed, perRequest)
	logger.Printf("********************************")

	result := getData(c)
	fileName := fmt.Sprintf("./perfresult-follower-%s", *followerID)
	err := ioutil.WriteFile(fileName, []byte(result), 0644)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	configuration := config.ReadConfiguration()

	switch *logOutput {
	case "stdout-only":
		logger = log.New(os.Stdout, "", log.LstdFlags)
	case "both":
		logFile, err := os.OpenFile(fmt.Sprintf("c-%s.log", *followerID),
			os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println(err)
		}
		defer logFile.Close()

		logger = log.New(logFile, "", log.LstdFlags)

		mw := io.MultiWriter(os.Stdout, logFile)
		logger.SetOutput(mw)
	default:
		// Write to log file only
		logFile, err := os.OpenFile(fmt.Sprintf("c-%s.log", *followerID),
			os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println(err)
		}
		defer logFile.Close()

		logger = log.New(logFile, "", log.LstdFlags)
	}

	// There must be a follower specified via flag, and that follower must be in
	// the config.
	followerAddress, ok := configuration.Followers[*proto.String(*followerID)]
	if !ok {
		logger.Fatalf("did not specify a follower ID with the %q flag", followerIDFlag)
	}
	// Connect to the follower.
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(time.Duration(*timeout)*time.Second))
	logger.Printf("Start client connecting to follower %q at address %q", *followerID, followerAddress)
	// Only supporting localhost for testing.
	conn, err := grpc.Dial(util.FormatServiceAddress(followerAddress), opts...)
	if err != nil {
		logger.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := flpb.NewFollowerClient(conn)
	logger.Printf("client mode is %v", *mode)
	switch *mode {
	case "normal":
		runClient(client)
	case "perf":
		perfTest(client)
	}

}
