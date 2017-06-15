package resolver

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"fmt"
)

func SetupRemoteContainerLink(peerIp string, linkId graph.LinkID, tunnelId uint64) error {
	fmt.Println("Dummy setup", linkId, "to", peerIp, "via", tunnelId)
	return nil
}

func TeardownRemoteContainerLink(peerIp string, linkId graph.LinkID) error {
	fmt.Println("Dummy teardown", linkId, "to", peerIp)
	return nil
}
