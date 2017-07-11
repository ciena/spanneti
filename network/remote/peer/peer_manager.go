package peer

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"fmt"
	"sync"
)

type PeerManager struct {
	mutex sync.Mutex
	peer  map[string]*remotePeer
}

func NewManager() PeerManager {
	return PeerManager{
		peer: make(map[string]*remotePeer),
	}
}

func (man *PeerManager) TryAllocate(linkId graph.LinkID, ethName string, containerPid int, tunnelId TunnelID, fabricIp string) (bool, error) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	//check if this tunnelId is in use (it's OK if we it's already used for this peer & tunnelId)
	for _, peer := range man.peer {
		if currentLinkId, have := peer.linkFor[tunnelId]; have && (currentLinkId != linkId || peer.fabricIp != fabricIp) {
			return false, nil
		}
	}

	//if this line is reached, it is valid to allocate

	peer := man.getPeerUnsafe(fabricIp)
	if err := peer.allocate(linkId, ethName, containerPid, tunnelId, fabricIp); err != nil {
		return false, err
	}
	return true, nil
}

func (man *PeerManager) Deallocate(fabricIp string, linkId graph.LinkID, ethName string, containerPid int) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	peer := man.getPeerUnsafe(fabricIp)
	if err := peer.deallocate(linkId, ethName, containerPid); err != nil {
		fmt.Println(err)
	}
	return nil
}

func (man *PeerManager) FindExisting(linkId graph.LinkID, ethName string, containerPid int) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	fabricIp, tunnelId, err := resolver.FindExisting(ethName, containerPid)
	if err != nil {
		return err
	}
	if tunnelId != nil {
		peer := man.getPeerUnsafe(fabricIp)
		if err := peer.allocate(linkId, ethName, containerPid, TunnelID(*tunnelId), fabricIp); err != nil {
			return err
		}
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

func (man *PeerManager) TunnelFor(fabricIp string, linkId graph.LinkID) (*TunnelID, bool) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	peer, have := man.peer[fabricIp]
	if !have {
		return nil, false
	}
	tunnelId, have := peer.tunnelFor[linkId]
	return &tunnelId, have
}

func (man PeerManager) getPeerUnsafe(fabricIp string) *remotePeer {
	if peer, have := man.peer[fabricIp]; have {
		return peer
	} else {
		peer := &remotePeer{
			fabricIp:  fabricIp,
			tunnelFor: make(map[graph.LinkID]TunnelID),
			linkFor:   make(map[TunnelID]graph.LinkID),
		}
		man.peer[fabricIp] = peer
		return peer
	}
}
