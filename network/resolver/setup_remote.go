package resolver

import (
	"fmt"
	"github.com/khagerma/cord-networking/network/graph"
)

func SetupRemoteContainerLink(peerIp string, linkId graph.LinkID, tunnelId uint64) error {
	fmt.Println("Dummy setup", tunnelId, "at", linkId, "to", peerIp)
	return nil
}

func TeardownRemoteContainerLink(peerIp string, linkId graph.LinkID) error {
	fmt.Println("Dummy teardown", linkId, "to", peerIp)
	return nil
}
