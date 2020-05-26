package util

import (
	"fmt"

	cpb "distributed-key-value-store/protos/config"
)

// FormatServiceAddress returns the address formatted for Dial and similar
// functions.
func FormatServiceAddress(node *cpb.ServiceAddress) string {
	return fmt.Sprintf("%s:%d", node.Address, node.Port)
}
