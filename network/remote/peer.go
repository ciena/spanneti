package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"fmt"
	"net"
	"os"
	"sync"
)

type peerID string
type tunnelID uint32

const MAX_TUNNEL_ID = 16777216
const DNS_ENTRY = "%s.%s.svc.cluster.local"

var SERVICE = os.Getenv("SERVICE")
var NAMESPACE = os.Getenv("NAMESPACE")

type remotePeer struct {
	peerId    peerID
	fabricIp  string
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
		if err := resolver.TeardownRemoteContainerLink(peer.fabricIp, linkId); err != nil {
			return err
		}
	}

	//map new relationship
	peer.tunnelFor[linkId] = tunnelId
	peer.linkFor[tunnelId] = linkId

	fmt.Printf("Will setup link %s to %s(%s) via %d\n", linkId, peer.fabricIp, peer.peerId, tunnelId)
	err := resolver.SetupRemoteContainerLink(peer.fabricIp, linkId, uint32(tunnelId))
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
		return resolver.TeardownRemoteContainerLink(peer.fabricIp, linkId)
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

func (man *RemoteManager) getPeer(peerId peerID) (*remotePeer, error) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	if peer, have := man.peer[peerId]; have {
		return peer, nil
	} else {
		info, err := man.requestInfo(peerId)
		if err != nil {
			return nil, err
		}

		peer := &remotePeer{
			peerId:    peerId,
			fabricIp:  info.FabricIp,
			tunnelFor: make(map[graph.LinkID]tunnelID),
			linkFor:   make(map[tunnelID]graph.LinkID),
		}
		man.peer[peerId] = peer
		return peer, nil
	}
}

func lookupPeerIps() ([]peerID, error) {
	ips, err := net.LookupIP(fmt.Sprintf(DNS_ENTRY, SERVICE, NAMESPACE))
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
