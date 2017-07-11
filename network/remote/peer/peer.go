package peer

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"fmt"
)

type PeerID string
type TunnelID uint32

const NUM_TUNNEL_IDS = 1 << 24

type remotePeer struct {
	fabricIp  string
	tunnelFor map[graph.LinkID]TunnelID
	linkFor   map[TunnelID]graph.LinkID
}

func (peer *remotePeer) allocate(linkId graph.LinkID, ethName string, containerPid int, tunnelId TunnelID, fabricIp string) error {
	//cleanup old relationships
	if oldTunnelId, have := peer.tunnelFor[linkId]; have {
		if oldTunnelId == tunnelId {
			fmt.Println("Alreay set up:", linkId, "to", fabricIp, "via", tunnelId)
			return nil
		}
		delete(peer.linkFor, oldTunnelId)
		delete(peer.tunnelFor, linkId)
		if err := resolver.TeardownRemoteContainerLink(ethName, containerPid); err != nil {
			return err
		}
	}

	//map new relationship
	peer.tunnelFor[linkId] = tunnelId
	peer.linkFor[tunnelId] = linkId

	fmt.Printf("Will setup link %s:%s to %s via %d\n", ethName, linkId, fabricIp, tunnelId)
	err := resolver.SetupRemoteContainerLink(ethName, containerPid, int(tunnelId), fabricIp)
	if err != nil {
		delete(peer.tunnelFor, linkId)
		delete(peer.linkFor, tunnelId)
	}
	return err
}

func (peer *remotePeer) deallocate(linkId graph.LinkID, ethName string, containerPid int) error {
	//cleanup old relationships
	if tunnelId, have := peer.tunnelFor[linkId]; have {
		delete(peer.linkFor, tunnelId)
		delete(peer.tunnelFor, linkId)
		return resolver.TeardownRemoteContainerLink(ethName, containerPid)
	}
	return nil
}
