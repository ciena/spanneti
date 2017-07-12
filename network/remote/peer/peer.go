package peer

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"errors"
	"fmt"
)

type PeerID string
type TunnelID uint32

const NUM_TUNNEL_IDS = 1 << 24

func (man *TunnelManager) allocate(linkId graph.LinkID, ethName string, containerPid int, tunnelId TunnelID, fabricIp string) error {
	//cleanup old relationships
	if tunnel, have := man.tunnelForLink[linkId]; have {
		if tunnel.id == tunnelId && tunnel.fabricIp == fabricIp {
			fmt.Printf("Alreay setup link %s:%s to %s via %d\n", ethName, linkId, fabricIp, tunnelId)
			return nil
		}
		return errors.New(fmt.Sprint("Tunnel with linkId", linkId, " already exists"))
	}

	if _, have := man.tunnelForId[tunnelId]; have {
		return errors.New(fmt.Sprint("Tunnel with tunnelId ", tunnelId, " already exists"))
	}

	fmt.Printf("Will setup link %s:%s to %s via %d\n", ethName, linkId, fabricIp, tunnelId)
	if err := resolver.SetupRemoteContainerLink(ethName, containerPid, int(tunnelId), fabricIp); err != nil {
		return err
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
	return nil
}

func (man *TunnelManager) deallocate(linkId graph.LinkID) error {
	//cleanup old relationships
	if tunnel, have := man.tunnelForLink[linkId]; have {
		fmt.Printf("Will teardown link %s:%s to %s via %d\n", tunnel.ethName, linkId, tunnel.fabricIp, tunnel.id)
		if _, _, err := resolver.DeleteContainerRemoteInterface(tunnel.ethName, tunnel.containerPid); err != nil {
			return err
		}
		delete(man.tunnelForId, tunnel.id)
		delete(man.tunnelForLink, linkId)
	}
	return nil
}
