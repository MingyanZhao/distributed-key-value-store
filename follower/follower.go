package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"distributed-key-value-store/config"
	configpb "distributed-key-value-store/protos/config"
	pb "distributed-key-value-store/protos/follower"
	lpb "distributed-key-value-store/protos/leader"
	"distributed-key-value-store/util"
)

const followerIDFlag = "follower_id"
const timeoutFlag = "timeout"

// Flags.
var (
	followerID            = flag.String(followerIDFlag, "", "The Follower's Unique ID")
	timeout               = flag.Int(timeoutFlag, 5, "Timeout, in seconds, when connecting to the leader")
	followerConnectionMap = make(map[string]pb.FollowerClient)
)

// follower implements the Follower service.
type follower struct {
	// TODO: Remove. This stubs the Follower methods we haven't implemented yet.
	// Probably should keep it. We get Unimplemented error for free.
	pb.UnimplementedFollowerServer

	// These are set during creation and are immutable.
	id      string
	address *configpb.ServiceAddress
	done    chan bool // TODO: Unused.

	// TODO: What happens if we lose the connection to the leader?
	// Need to handle it. Probably need to take a look at https://github.com/grpc/grpc-go/tree/master/examples/features/keepalive
	leader lpb.LeaderClient
	conn   *grpc.ClientConn

	// TODO: store needs a mutex.
	store map[string]*data
}

type data struct {
	// versionToValues maintains all of the versionToValues that has been accepted by the leader
	// version to values mapping is convenient for implementation
	versionToValues    map[int64][]string
	versionToValueLock sync.RWMutex

	// latest version of the key
	latestVersion int64

	// Unpdated values are the input values that has not been accepted by the leader
	buffer []string

	// Each key has a update channel, so that different keys are updated concurrently.
	// The values for the same key are not updated in parallel because of linearizable consideration.
	updateReqChan chan *lpb.UpdateRequest

	// syncReqChan is the channel that sending the sync requests to the target followers
	syncReqChan  chan *syncRequest
	syncReqQueue []*syncRequest

	// syncRespChan is the channel that handles the received sync responses.
	syncRespChan chan *pb.SyncResponse
	done         chan bool
}

type value struct {
	val     string
	version int64
}

// syncRequest wraps the necessary information for sending a sync request
type syncRequest struct {
	req         *pb.SyncRequest
	primaryID   string
	primaryAddr *configpb.ServiceAddress
	backupID    string
	backupAddr  *configpb.ServiceAddress
}

func dial(addr *configpb.ServiceAddress) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(time.Duration(*timeout)*time.Second))
	conn, err := grpc.Dial(util.FormatServiceAddress(addr), opts...)
	if err != nil {
		return nil, fmt.Errorf("fail to dial server at address %v: %v", addr, err)
	}
	return conn, nil
}

// Create a follower that is connected to the leader.
func newFollower(configuration *configpb.Configuration, followerID string, followerAddress *configpb.ServiceAddress) (*follower, error) {
	// Connect to the leader
	conn, err := dial(configuration.Leader)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to leader at address %v", configuration.Leader)
	return &follower{
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

func (f *follower) initNewKey(key string) {
	// A new key is received. Initialize a new entry in the store.
	f.store[key] = &data{
		updateReqChan:   make(chan *lpb.UpdateRequest, 100),
		syncReqChan:     make(chan *syncRequest, 100),
		syncReqQueue:    make([]*syncRequest, 0),
		done:            make(chan bool),
		buffer:          make([]string, 0),
		versionToValues: make(map[int64][]string),
	}
	// Starts channels
	go f.handleUpdate(key)
	go f.handleSync(key)
}

func (f *follower) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	// Storing the data
	log.Printf("Received Put request %v", req)

	// Get the current version (or 0 if it's the first).
	if _, ok := f.store[req.Key]; !ok {
		f.initNewKey(req.Key)
	}

	// newData := &value{val: req.Value, version: v}

	// The incoming values are put in the buffer and waiting to be udpated by the leader.
	f.store[req.Key].buffer = append(f.store[req.Key].buffer, req.Value)
	// f.store[req.Key].values = append(f.store[req.Key].values, newData)

	res := fmt.Sprintf("key value pair recieved %v: %v", req.Key, req.Value)
	log.Println(res)

	// TODO: write the data to the local disk
	// Infomation that needs to be written:
	// key, value, state
	// The states are:
	// received, updated, synced

	go f.sendUpdate(req.Key)

	// Reply to the clilent
	return &pb.PutResponse{
		Result: res,
	}, nil
}

// TODO: A slightly better impl. May need more refactoring.
func (f *follower) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	log.Printf("Received a Get request %v", req)
	var result []string
	// First get the data from the store, which are all of the synced data.
	if d, ok := f.store[req.Key]; ok {
		for _, v := range d.versionToValues {
			result = append(result, v...)

			// then append all of the data that still waiting for update.

		}
		result = append(result, f.store[req.Key].buffer...)
	}

	log.Printf("Get result is: [%v]", result)

	return &pb.GetResponse{
		Values: result,
	}, nil
}

func (f *follower) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	log.Printf("Received Sync request %v", req)
	if _, ok := f.store[req.Key]; !ok {
		return nil, fmt.Errorf("sync message discarded because the key %v is not found", req.Key)
	}

	askForVers := req.AskFor.Versions
	incomingData := req.MyData
	var outValues []*pb.Value
	mydata := f.store[req.Key].versionToValues
	// The version numbers should all exist in the store.
	// Only reply the versions that being asked for.
	for _, v := range askForVers {
		if _, ok := mydata[v]; !ok {
			err := fmt.Errorf("reply to sync failed, missing version %v for key %v, request %v", v, req.Key, req)
			log.Print(err)
			return nil, err
		}

		outValues = append(outValues, &pb.Value{
			Value:   mydata[v],
			Version: v,
		})
	}

	// Also cosume the incoming data
	if incomingData != nil {
		mydata[incomingData.Version] = incomingData.Values
		// Incoming version should always be the latest version
		log.Printf("current latest version is %v, incoming version is %v", f.store[req.Key].latestVersion, incomingData.Version)
		f.store[req.Key].latestVersion = incomingData.Version
	}

	return &pb.SyncResponse{
		Key:   req.Key,
		Value: outValues,
	}, nil
}

func (f *follower) sendUpdate(key string) {
	// TODO: Make this happen asynchronously.
	// Sending update with leader
	log.Println("update with global leader...")

	updateReq := &lpb.UpdateRequest{
		Key:             key,
		FollowerAddress: f.address,
		Version:         f.store[key].latestVersion,
		FollowerId:      *proto.String(f.id),
	}
	f.store[key].updateReqChan <- updateReq
}

func (f *follower) sendSync(data *data, req *syncRequest) {
	data.syncReqQueue = append(data.syncReqQueue, req)
}

func (f *follower) handleSync(key string) error {
	log.Printf("start sync handler for key %v", key)
	syncReqChan := f.store[key].syncReqChan
	ctx := context.Background()
	for {
		select {
		case syncReqest := <-syncReqChan:
			log.Printf("handling sync request id %v, addr %v", syncReqest.primaryID, *syncReqest.primaryAddr)
			// Sending the sync request to the target follower one by one, then put into the syncResp channel.
			// TODO: This is a blocking procedure, maybe able to change to a concurrent procedure
			// Record the connection after dailing. Only dail on the first time.
			if _, ok := followerConnectionMap[syncReqest.primaryID]; !ok {
				newConn, err := dial(syncReqest.primaryAddr)
				if err != nil {
					// TODO: handle error
					log.Printf("error: failed to connect the pre-primary %v, at address: %v", syncReqest.primaryID, syncReqest.primaryAddr)
					continue
				}
				followerConnectionMap[syncReqest.primaryID] = pb.NewFollowerClient(newConn)
			}
			prePrimary := followerConnectionMap[syncReqest.primaryID]

			// Call sync on the target follower
			syncResp, err := prePrimary.Sync(ctx, syncReqest.req)
			if err != nil {
				// TODO: handle error
				log.Printf("Sync rpc failed, %v", err)
				continue
			}

			// syncResp received, update local store with incoming data
			log.Printf("handling sync response, %v", syncResp)
			newValues := syncResp.Value
			f.store[key].versionToValueLock.Lock()
			myData := f.store[key].versionToValues
			// Put each value list under its version.
			for _, v := range newValues {
				if d, ok := myData[v.Version]; ok {
					err := fmt.Errorf("something went wrong, the version number should not be conflicting, key %v, existing data %v, incomming data %v", key, d, v)
					log.Println(err)
					continue
				}
				myData[v.Version] = v.Value
			}
			f.store[key].versionToValueLock.Unlock()

			log.Printf("SyncResp done, %v", syncResp)
		}
	}
}

// askForVersions generates a list of versions in an open range of (localVer, globalVer)
func askForVersions(globalVer, localVer int64) *pb.AskFor {
	result := &pb.AskFor{}
	var i int64
	for i = localVer + 1; i < globalVer; i++ {
		result.Versions = append(result.Versions, i)
	}
	return result
}

func (f *follower) handleUpdate(key string) error {
	log.Printf("start update handler for key %v", key)
	ctx := context.Background()
	done := f.store[key].done
	updateReqChan := f.store[key].updateReqChan
	for {
		select {
		case <-done:
			log.Printf("key %v update channel closed", key)
			return nil
		case updateReq := <-updateReqChan:
			// there is an update request
			log.Printf("sending update request %v", updateReq)
			updateResp, err := f.leader.Update(ctx, updateReq)
			if err != nil {
				// TODO: handle the error properlly.
				// One update message failed but still need to continue.
				log.Printf("failed %v", err)
			}
			log.Printf("handling update response %v for key %v", updateResp, key)
			data := f.store[key]
			globalVer := updateResp.Version
			localVer := data.latestVersion
			// The input argument globalVer below means not asking for the latest version, but the missing version between globalVer and localVer.
			askForVers := askForVersions(globalVer, localVer)

			// Upon receiving an update response, move the buffered data data to commited values
			// and updat the latest version, since at this point the leader has accepted the proposed value.
			f.store[key].versionToValueLock.Lock()
			data.versionToValues[globalVer] = append(data.versionToValues[globalVer], data.buffer...)
			f.store[key].versionToValueLock.Unlock()

			if updateResp.Result == lpb.UpdateResult_NEED_SYNC {
				syncReq := &pb.SyncRequest{
					Key:    key,
					AskFor: askForVers,
					MyData: &pb.Mydata{
						Version: updateResp.Version,
						Values:  data.buffer,
					},
				}
				s := &syncRequest{
					req:         syncReq,
					primaryID:   updateResp.PrePrimary.FollowerId,
					primaryAddr: updateResp.PrePrimary.Address,
					backupID:    updateResp.PreBackup.FollowerId,
					backupAddr:  updateResp.PreBackup.Address,
				}
				go func() {
					data.syncReqChan <- s
				}()
				log.Printf("sync request is in the channel")
			}
			data.buffer = data.buffer[:0]
			data.latestVersion = updateResp.Version
			log.Printf("key %v needs a sync? %v", key, updateResp.Result)
		}
	}
}

// TODO: Handle.
func (f *follower) Notify(ctx context.Context, req *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	log.Printf("Got a notification, %v", req)
	if _, ok := f.store[req.Key]; !ok {
		// Got a new Key
		f.initNewKey(req.Key)
	}

	data := f.store[req.Key]
	localVer := data.latestVersion
	if localVer == req.Version {
		// No need to update
		return &pb.NotifyResponse{Success: true}, nil
	}

	// The global versioin argument below needs a +1, which means also asking for the latest version on the target follower.
	// It does not need to ask for the local version either since it already has it.
	askForVers := askForVersions(req.Version+1, localVer)
	syncReq := &pb.SyncRequest{
		Key:    req.Key,
		AskFor: askForVers,
		// Not carrying mydata, since it is expected to be outdated.
	}
	s := &syncRequest{
		req:         syncReq,
		primaryID:   req.Primary.FollowerId,
		primaryAddr: req.Primary.Address,
		backupID:    req.Backup.FollowerId,
		backupAddr:  req.Backup.Address,
	}
	log.Printf("sync request is in the channel, %v", s)
	go func() {
		data.syncReqChan <- s
	}()

	// TODO: Need to think about if this is safe.
	// The version is updated after handling the braodcasting notification.
	// SInce we have asked for till the req.Version+1, the local version needs to be updated to that.
	data.latestVersion = req.Version

	return &pb.NotifyResponse{
		Success: true,
	}, nil
}

func main() {
	flag.Parse()

	// Get config and options.
	configuration := config.ReadConfiguration()
	followerAddress, ok := configuration.Followers[*proto.String(*followerID)]
	if !ok {
		log.Fatalf("did not specify a follower ID with the %q flag", followerIDFlag)
	}

	// Create and start the follower service.
	lis, err := net.Listen("tcp", util.FormatServiceAddress(followerAddress))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	f, err := newFollower(configuration, *followerID, followerAddress)
	if err != nil {
		log.Fatalf("failed to create new follower: %v", err)
	}

	log.Printf("Starting follower server and listening on address %v", followerAddress)
	grpcServer := grpc.NewServer()
	pb.RegisterFollowerServer(grpcServer, f)
	grpcServer.Serve(lis)
}
