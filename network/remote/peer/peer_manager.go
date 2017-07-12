package peer

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"fmt"
	"sync"
)

type TunnelManager struct {
	mutex sync.Mutex
	//peer             map[string]*remotePeer
	tunnelForId   map[TunnelID]*tunnel
	tunnelForLink map[graph.LinkID]*tunnel
}

type tunnel struct {
	id       TunnelID
	linkId   graph.LinkID
	fabricIp string
}

func NewManager() TunnelManager {
	return TunnelManager{
		tunnelForId:   make(map[TunnelID]*tunnel),
		tunnelForLink: make(map[graph.LinkID]*tunnel),
	}
}

func (man *TunnelManager) TryAllocate(linkId graph.LinkID, ethName string, containerPid int, tunnelId TunnelID, fabricIp string) (bool, error) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	//check if this tunnelId is in use ()
	if tunnel, have := man.tunnelForId[tunnelId]; have{
		//it's OK if it's already used for this peer & tunnelId
		if tunnel.linkId != linkId || tunnel.fabricIp != fabricIp {
			return false, nil
		}
	}

	//if this line is reached, it is valid to allocate

	if err := man.allocate(linkId, ethName, containerPid, tunnelId, fabricIp); err != nil {
		return false, err
	}
	return true, nil
}

func (man *TunnelManager) Deallocate(linkId graph.LinkID, ethName string, containerPid int) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	if err := man.deallocate(linkId, ethName, containerPid); err != nil {
		fmt.Println(err)
	}
	return nil
}

func (man *TunnelManager) FindExisting(linkId graph.LinkID, ethName string, containerPid int) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	fabricIp, tunnelId, exists, err := resolver.FindExisting(ethName, containerPid)
	if err != nil {
		return err
	}
	if exists {
		tunnel := &tunnel{
			id:       TunnelID(tunnelId),
			linkId:   linkId,
			fabricIp: fabricIp,
		}
		man.tunnelForId[TunnelID(tunnelId)] = tunnel
		man.tunnelForLink[linkId] = tunnel
	}
	return nil
}

func (man *TunnelManager) NextAvailableTunnelId(after TunnelID) *TunnelID {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	for ; after < NUM_TUNNEL_IDS; after++ {
		if _, has := man.tunnelForId[after]; !has {
			return &after
		}
	}
	return nil
}

func (man *TunnelManager) TunnelFor(fabricIp string, linkId graph.LinkID) (TunnelID, bool) {
	tunnel, allocated := man.tunnelForLink[linkId]
	if !allocated || tunnel.fabricIp != fabricIp {
		return 0, false
	}
	return tunnel.id, allocated
}
