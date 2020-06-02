package util

import (
	"fmt"

	cpb "distributed-key-value-store/protos/config"

	proto "github.com/golang/protobuf/proto"
)

// FormatServiceAddress returns the address formatted for Dial and similar
// functions.
func FormatServiceAddress(node *cpb.ServiceAddress) string {
	return fmt.Sprintf("%s:%d", node.Address, node.Port)
}

// FormatBindAddress returns to address formatted for binding by changing
// the address to 0.0.0.0 if it is not locahost
func FormatBindAddress(node *cpb.ServiceAddress) string {
	bindAddress := proto.Clone(node).(*cpb.ServiceAddress)
	if bindAddress.Address != "localhost" {
		bindAddress.Address = "0.0.0.0"
	}
	return FormatServiceAddress(bindAddress)
}
