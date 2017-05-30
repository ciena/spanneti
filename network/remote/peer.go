package remote

import (
	"github.com/khagerma/cord-networking/network/graph"
	"sync"
	"github.com/khagerma/cord-networking/network/resolver"
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

func (peer *remotePeer) allocate(linkId graph.LinkID, tunnelId tunnelID) {
	//cleanup old relationships
	if oldTunnelId, have := peer.tunnelFor[linkId]; have {
		delete(peer.linkFor, oldTunnelId)
	}
	if oldLinkId, have := peer.linkFor[tunnelId]; have {
		delete(peer.tunnelFor, oldLinkId)
	}

	//map new relationship
	peer.tunnelFor[linkId] = tunnelId
	peer.linkFor[tunnelId] = linkId

	resolver.SetupRemoteContainerLink(string(peer.peerId), linkId, uint64(tunnelId))
}

func (peer *remotePeer) nextAvailableTunnelId(after tunnelID) *tunnelID {
	for ; after < MAX_TUNNEL_ID; after++ {
		if _, has := peer.linkFor[after]; !has {
			return &after
		}
	}
	return nil
}
