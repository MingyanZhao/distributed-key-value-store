package config

import (
	"flag"
	"io/ioutil"
	"log"

	pb "distributed-key-value-store/protos/config"

	"github.com/golang/protobuf/proto"
)

var (
	configFilePath = flag.String("config_file", "config/local-config.pb.txt", "The Configuration file")
)

// ReadConfiguration  -- Read a Configuration from the --config_file
func ReadConfiguration() *pb.Configuration {

	content, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		log.Fatalf("Failed to read configuration file %v", err)
	}

	configuration := &pb.Configuration{}
	if err := proto.UnmarshalText(string(content), configuration); err != nil {
		log.Fatalf("Failed to parse configuration file %v", err)
	}

	return configuration
}
