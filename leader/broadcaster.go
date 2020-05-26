package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	cpb "distributed-key-value-store/protos/config"
	fpb "distributed-key-value-store/protos/follower"
	pb "distributed-key-value-store/protos/leader"
)

var frequencyMS = flag.Int("frequency", 1000, "The frequency of periodic broadcasts in milliseconds")

type broadcaster interface {
	// Asynchrounously enqueue and update to be broadcasted. When it is sent
	// depends in the implementer - it is only guaranteed to be sent eventually.
	enqueue(key string, resp *pb.UpdateResponse)
}

// immediateCaster asynchronously broadcasts updates as soon as possible. Useful
// for debugging and latency-sensitive data.
//
// Should be created with newImmediate.
type immediateCaster struct {
	notifier
}

func newImmediate(config *cpb.Configuration) (*immediateCaster, error) {
	notifier, err := newNotifier(config)
	if err != nil {
		return nil, err
	}

	return &immediateCaster{
		notifier: notifier,
	}, nil
}

func (ic *immediateCaster) enqueue(key string, resp *pb.UpdateResponse) {
	ic.notify(key, resp)
}

// periodicCaster broadcasts at a regular interval.
//
// Should be created with newPeriodic.
type periodicCaster struct {
	notifier
	// TODO: Unused.
	done chan bool
	// Updates holds the list of updates that will be broadcast at the next
	// interval. Rather than a buffered channel of updates, a channel holding a
	// slice lets the slice grow unbounded.
	updates chan []update
}

type update struct {
	key  string
	resp *pb.UpdateResponse
}

func newPeriodic(config *cpb.Configuration) (*periodicCaster, error) {
	// Connect to each follower.
	notifier, err := newNotifier(config)
	if err != nil {
		return nil, err
	}

	pc := periodicCaster{
		notifier: notifier,
		done:     make(chan bool),
		updates:  make(chan []update, 1),
	}
	pc.updates <- nil

	// Kick off a goroutine that wakes up periodically and sends updates.
	go func() {
		ticker := time.NewTicker(time.Duration(*frequencyMS) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-pc.done:
				log.Printf("Periodic broadcaster done.")
				return
			case <-ticker.C:
				// This has a bit of a "stop-the-world" problem, as incoming updates
				// will block until we launch len(updates) goroutines.
				updates := <-pc.updates
				for _, update := range updates {
					pc.notify(update.key, update.resp)
				}
				pc.updates <- nil
			}
		}
	}()

	return &pc, nil
}

func (pc *periodicCaster) enqueue(key string, resp *pb.UpdateResponse) {
	updates := <-pc.updates
	updates = append(updates, update{key: key, resp: resp})
	pc.updates <- updates
}

// notifier knows how to connect to and notify all the followers in a config.
//
// Should only be created with newNotifier.
type notifier struct {
	// followers maps a follower ID to it's GRPC client.
	// TODO: Call Close() on all clients.
	followers map[string]fpb.FollowerClient
}

func newNotifier(config *cpb.Configuration) (notifier, error) {
	// Establish a GRPC client for each follower.
	notifier := notifier{
		followers: make(map[string]fpb.FollowerClient),
	}
	for id, addr := range config.FollowerAddresses {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return notifier, fmt.Errorf("failed to dial follower %q at address %q: %v", id, addr, err)
		}
		notifier.followers[id] = fpb.NewFollowerClient(conn)
	}
	return notifier, nil
}

// Asynchrounously send an RPC to every follower (except the one the response is
// already being sent to).
func (nt notifier) notify(key string, resp *pb.UpdateResponse) {
	for id, client := range nt.followers {
		// We don't have to broadcast the update to the updater.
		if resp.PrePrimary.FollowerId == id {
			continue
		}
		go func(id string, client fpb.FollowerClient) {
			req := fpb.NotifyRequest{
				Key: key,
			}
			// TODO: Handle.
			_, err := client.Notify(context.Background(), &req)
			if err != nil {
				log.Printf("error notifying follower %q", id)
				return
			}
			log.Printf("notified follower %q", id)
		}(id, client)
	}

}
