package link

import (
	"fmt"
	"github.com/ciena/spanneti/resolver"
	"sync"
)

type tunnelID uint32

const NUM_TUNNEL_IDS = 1 << 24

type tunnelManager struct {
	mutex sync.Mutex
	//peer             map[string]*remotePeer
	tunnelForId   map[tunnelID]*tunnel
	tunnelForLink map[linkID]*tunnel
}

type tunnel struct {
	id           tunnelID
	linkId       linkID
	fabricIp     string
	ethName      string
	containerPid int
}

func newTunnelManager() tunnelManager {
	return tunnelManager{
		tunnelForId:   make(map[tunnelID]*tunnel),
		tunnelForLink: make(map[linkID]*tunnel),
	}
}

func (man *tunnelManager) allocate(linkId linkID, ethName string, containerPid int, tunnelId tunnelID, fabricIp string) (bool, error) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	//if this tunnelId is in use
	if tunnel, have := man.tunnelForId[tunnelId]; have {
		//if linkId is incorrect, fail
		if tunnel.linkId != linkId {
			return false, nil
		}

		//if everything else is correct (fabricIp, ethName, containerPid), nothing to do
		if tunnel.fabricIp == fabricIp && tunnel.ethName == ethName && tunnel.containerPid == containerPid {
			fmt.Printf("Existing link %s:%s to %s via %d\n", ethName, linkId, fabricIp, tunnelId)
			return true, nil
		}

		//if anything else is incorrect, reallocate
		if err := man.deallocateUnsafe(linkId); err != nil {
			return false, err
		}
	}

	fmt.Printf("Setup link %s:%s to %s via %d\n", ethName, linkId, fabricIp, tunnelId)
	if err := resolver.SetupRemoteContainerLink(ethName, containerPid, int(tunnelId), fabricIp); err != nil {
		if err.Error() == "tunnel unavailable" {
			fmt.Println("TunnelId", tunnelId, "unavailable")
			return false, nil
		}
		return false, err
	}

	//map new relationship
	tunnel := &tunnel{
		id:           tunnelId,
		linkId:       linkId,
		fabricIp:     fabricIp,
		ethName:      ethName,
		containerPid: containerPid,
	}
	man.tunnelForId[tunnelId] = tunnel
	man.tunnelForLink[linkId] = tunnel
	return true, nil
}

func (man *tunnelManager) deallocate(linkId linkID) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	return man.deallocateUnsafe(linkId)
}

func (man *tunnelManager) deallocateUnsafe(linkId linkID) error {
	if tunnel, have := man.tunnelForLink[linkId]; have {
		fmt.Printf("Teardown link %s:%s to %s via %d\n", tunnel.ethName, linkId, tunnel.fabricIp, tunnel.id)
		if err := resolver.DeleteContainerRemoteInterface(tunnel.ethName, tunnel.containerPid); err != nil {
			return err
		}
		delete(man.tunnelForId, tunnel.id)
		delete(man.tunnelForLink, linkId)
	}
	return nil
}

func (man *tunnelManager) findExisting(linkId linkID, ethName string, containerPid int) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	fabricIp, tunnelId, exists, err := resolver.FindExistingRemoteInterface(ethName, containerPid)
	if err != nil {
		return err
	}
	if exists {
		tunnel := &tunnel{
			id:           tunnelID(tunnelId),
			linkId:       linkId,
			fabricIp:     fabricIp,
			ethName:      ethName,
			containerPid: containerPid,
		}
		man.tunnelForId[tunnelID(tunnelId)] = tunnel
		man.tunnelForLink[linkId] = tunnel
	}
	return nil
}

func (man *tunnelManager) firstAvailableTunnelId() *tunnelID {
	return man.thisOrNextAvailableTunnelId(0)
}

func (man *tunnelManager) nextAvailableTunnelId(after tunnelID) *tunnelID {
	return man.thisOrNextAvailableTunnelId(after + 1)
}

func (man *tunnelManager) thisOrNextAvailableTunnelId(tunnelId tunnelID) *tunnelID {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	for ; tunnelId < NUM_TUNNEL_IDS; tunnelId++ {
		if _, has := man.tunnelForId[tunnelId]; !has {
			return &tunnelId
		}
	}
	return nil
}

func (man *tunnelManager) tunnelFor(fabricIp string, linkId linkID) (tunnelID, bool) {
	tunnel, allocated := man.tunnelForLink[linkId]
	if !allocated || tunnel.fabricIp != fabricIp {
		return 0, false
	}
	return tunnel.id, allocated
}
