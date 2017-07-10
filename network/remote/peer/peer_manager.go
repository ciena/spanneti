package peer

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"fmt"
	"sync"
)

type PeerManager struct {
	mutex sync.Mutex
	peer  map[PeerID]*remotePeer
}

func NewManager() PeerManager {
	return PeerManager{
		peer: make(map[PeerID]*remotePeer),
	}
}

func (man *PeerManager) TryAllocate(peerId PeerID, linkId graph.LinkID, ethName string, containerPid int, tunnelId TunnelID, fabricIp string) (bool, error) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	//check if this tunnelId is in use (it's OK if we it's already used for this peer & tunnelId)
	for _, peer := range man.peer {
		if currentLinkId, have := peer.linkFor[tunnelId]; have && (currentLinkId != linkId || peer.peerId != peerId) {
			return false, nil
		}
	}

	//if this line is reached, it is valid to allocate

	peer := man.getPeerUnsafe(peerId)
	if err := peer.allocate(linkId, ethName, containerPid, tunnelId, fabricIp); err != nil {
		return false, err
	}
	return true, nil
}

func (man *PeerManager) Deallocate(peerId PeerID, linkId graph.LinkID) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	peer := man.getPeerUnsafe(peerId)
	if err := peer.deallocate(linkId); err != nil {
		fmt.Println(err)
	}
	return nil
}

func (man *PeerManager) NextAvailableTunnelId(after TunnelID) *TunnelID {
	man.mutex.Lock()
	defer man.mutex.Unlock()

outer:
	for ; after < NUM_TUNNEL_IDS; after++ {
		for _, peer := range man.peer {
			if _, has := peer.linkFor[after]; has {
				continue outer
			}
		}
		break
	}

	if after < NUM_TUNNEL_IDS {
		return &after
	}
	return nil
}

func (man *PeerManager) TunnelFor(peerId PeerID, linkId graph.LinkID) (*TunnelID, bool) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	peer, have := man.peer[peerId]
	if !have {
		return nil, false
	}
	tunnelId, have := peer.tunnelFor[linkId]
	return &tunnelId, have
}

func (man *PeerManager) getPeerUnsafe(peerId PeerID) *remotePeer {
	if peer, have := man.peer[peerId]; have {
		return peer
	} else {
		peer := &remotePeer{
			peerId:    peerId,
			tunnelFor: make(map[graph.LinkID]TunnelID),
			linkFor:   make(map[TunnelID]graph.LinkID),
		}
		man.peer[peerId] = peer
		return peer
	}
}
