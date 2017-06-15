package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"fmt"
	"net"
	"sync"
)

type peerID string
type tunnelID uint64

const MAX_TUNNEL_ID = 16777216
const DNS_ENTRY = "spanneti.default.svc.cluster.local"

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
			fmt.Println("Alreay set up:", linkId, "to", peer.peerId, "via", tunnelId)
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

func (man *RemoteManager) getPeer(peerId peerID) *remotePeer {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	if peer, have := man.peer[peerId]; have {
		return peer
	} else {
		peer := &remotePeer{
			peerId:    peerId,
			tunnelFor: make(map[graph.LinkID]tunnelID),
			linkFor:   make(map[tunnelID]graph.LinkID),
		}
		man.peer[peerId] = peer
		return peer
	}
}

func lookupPeerIps() ([]peerID, error) {
	ips, err := net.LookupIP(DNS_ENTRY)
	if err != nil {
		return []peerID{}, err
	}

	peers := []peerID{}
	for _, ip := range ips {
		if ip.To4() != nil {
			peers = append(peers, peerID(ip.String()))
		}
	}

	return peers, nil
}
