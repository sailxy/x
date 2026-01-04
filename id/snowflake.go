package id

import (
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	sfOnce   sync.Once
	sfNodeID int64
	sfNode   *snowflake.Node
	sfErr    error
)

func NewSnowflakeID() (int64, error) {
	n, err := snowflakeNode()
	if err != nil {
		return 0, err
	}
	id := n.Generate()
	return id.Int64(), nil
}

// snowflakeNode returns a stable node for the current machine/process.
//
// It is derived from machine identity:
// - /etc/machine-id (or /var/lib/dbus/machine-id) if present
// - otherwise hostname
//
// The node ID is in [0, 1023] to match the default 10-bit node space.
func snowflakeNode() (*snowflake.Node, error) {
	sfOnce.Do(func() {
		ident, err := machineIdentity()
		if err != nil {
			sfErr = err
			return
		}
		sfNodeID = int64(fnv32a(ident) % 1024)

		n, err := snowflake.NewNode(sfNodeID)
		if err != nil {
			sfErr = err
			return
		}
		sfNode = n
	})

	return sfNode, sfErr
}

// machineIdentity returns a stable machine identity for the current machine/process.
func machineIdentity() (string, error) {
	// Try stable machine ID first (Linux).
	for _, p := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		if b, err := os.ReadFile(p); err == nil {
			s := strings.TrimSpace(string(b))
			if s != "" {
				return s, nil
			}
		}
	}

	// Fallback: hostname.
	hn, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	hn = strings.TrimSpace(hn)
	if hn == "" {
		return "", fmt.Errorf("hostname is empty")
	}
	return hn, nil
}

// fnv32a is a hash function that is used to generate a stable node ID.
func fnv32a(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
