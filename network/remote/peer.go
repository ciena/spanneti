package remote

import (
	"fmt"
	"github.com/khagerma/cord-networking/network/graph"
	"github.com/khagerma/cord-networking/network/resolver"
	"sync"
)

type peerID string
type tunnelID uint64

const MAX_TUNNEL_ID = 16777216

type remotePeer struct {
	peerId    peerID
	tunnelFor map[graph.LinkID]tunnelID
	linkFor   map[tunnelID]graph.LinkID
	mutex     sync.Mutex
}

func (peer *remotePeer) allocate(linkId graph.LinkID, tunnelId tunnelID) error {
	//cleanup old relationships
	if oldTunnelId, have := peer.tunnelFor[linkId]; have {
		if oldTunnelId == tunnelId {
			fmt.Println("Alreay set up:", tunnelId, "at", linkId, "to", peer.peerId)
			return nil
		}
		delete(peer.linkFor, oldTunnelId)
		delete(peer.tunnelFor, linkId)
		if err := resolver.TeardownRemoteContainerLink(string(peer.peerId), linkId); err != nil {
			return err
		}
	}

	//map new relationship
	peer.tunnelFor[linkId] = tunnelId
	peer.linkFor[tunnelId] = linkId

	err := resolver.SetupRemoteContainerLink(string(peer.peerId), linkId, uint64(tunnelId))
	if err != nil {
		delete(peer.tunnelFor, linkId)
		delete(peer.linkFor, tunnelId)
	}
	return err
}

func (peer *remotePeer) deallocate(linkId graph.LinkID) error {
	//cleanup old relationships
	if tunnelId, have := peer.tunnelFor[linkId]; have {
		delete(peer.linkFor, tunnelId)
		delete(peer.tunnelFor, linkId)
		return resolver.TeardownRemoteContainerLink(string(peer.peerId), linkId)
	}
	return nil
}

func (peer *remotePeer) nextAvailableTunnelId(after tunnelID) *tunnelID {
	for ; after < MAX_TUNNEL_ID; after++ {
		if _, has := peer.linkFor[after]; !has {
			return &after
		}
	}
	return nil
}
