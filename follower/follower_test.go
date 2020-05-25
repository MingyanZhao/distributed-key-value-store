package main

import (
	"fmt"
	"testing"

	pb "distributed-key-value-store/protos/follower"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestAskForVersions(t *testing.T) {
	tests := []struct {
		desc      string
		globalVer int64
		localVer  int64
		want      *pb.AskFor
	}{
		{
			globalVer: 5,
			localVer:  1,
			want: &pb.AskFor{
				Versions: []int64{2, 3, 4},
			},
		},
		{
			globalVer: 2,
			localVer:  0,
			want: &pb.AskFor{
				Versions: []int64{1},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("global = %d, local= %d", test.globalVer, test.localVer), func(t *testing.T) {
			got := askForVersions(test.globalVer, test.localVer)
			if diff := cmp.Diff(test.want, got, protocmp.Transform()); diff != "" {
				t.Fatalf("processEgressRequest() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}
