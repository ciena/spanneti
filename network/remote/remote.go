package remote

import (
	"errors"
	"fmt"
	"github.com/khagerma/cord-networking/network/graph"
	"sync"
)

type RemoteManager struct {
	peerId      peerID
	peer        map[peerID]*remotePeer
	mutex       sync.Mutex
	resyncMutex sync.Mutex
	graph       *graph.Graph
	eventBus    chan<- graph.LinkID
	outOfSync   map[peerID]map[graph.LinkID]bool
}

func New(peerId string, network *graph.Graph, ch chan<- graph.LinkID) *RemoteManager {
	man := &RemoteManager{
		peerId:    peerID(peerId),
		peer:      make(map[peerID]*remotePeer),
		graph:     network,
		eventBus:  ch,
		outOfSync: make(map[peerID]map[graph.LinkID]bool),
	}

	go man.runServer()
	return man
}

//in order:
//get availability from remote
//provision on remote
//provision locally
func (man *RemoteManager) TryConnect(linkId graph.LinkID) (bool, error) {
	peerIps, possibilities := man.getPossibilities(linkId)

	if len(possibilities) != 1 {
		fmt.Println("Unable to find two containers with the same link ID.  Ignoring...")
		return false, nil
	}

	peerIp := peerIps[0]
	response := possibilities[0]

	tunnelId := response.TunnelId

	localSetup := false
	setup := response.Setup

	for !localSetup {
		for !setup {
			//
			// ensure the suggested tunnelId is valid on our side
			//

			var err error
			setup, tunnelId, err = man.requestSetup(peerIp, linkId, tunnelId)
			if err != nil {
				man.unableToSync(peerID(peerIp), linkId)
				return false, err
			}
		}

		//now that it's setup remotely, try to setup locally

		peer := man.getPeer(peerID(peerIp))
		peer.mutex.Lock()
		if currentLinkId, have := peer.linkFor[tunnelId]; !have || currentLinkId == linkId {
			err := peer.allocate(linkId, tunnelId)
			peer.mutex.Unlock()
			if err != nil {
				man.unableToSync(peerID(peerIp), linkId)
				return false, err
			}
			localSetup = true
		} else {
			//go to next available tunnel ID
			if tunnel := peer.nextAvailableTunnelId(tunnelId); tunnel == nil {
				man.unableToSync(peerID(peerIp), linkId)
				return false, errors.New("Out of tunnelIds?")
			} else {
				tunnelId = *tunnel
			}
			peer.mutex.Unlock()

			if _, err := man.requestDelete(peerIp, linkId); err != nil {
				man.unableToSync(peerID(peerIp), linkId)
				return false, err
			}
			setup = false
		}
	}
	return true, nil
}

func (man *RemoteManager) TryCleanup(linkId graph.LinkID) (deleted bool) {
	deleted = false
	for _, peerIp := range []string{"192.168.33.10:8080", "192.168.33.11:8080"} {
		if man.peerId == peerID(peerIp) {
			//do not connect to self
			continue
		}

		peer := man.getPeer(peerID(peerIp))
		peer.mutex.Lock()
		if err := peer.deallocate(linkId); err != nil {
			fmt.Println(err)
		}
		peer.mutex.Unlock()

		if wasDeleted, err := man.requestDelete(peerIp, linkId); err != nil {
			man.unableToSync(peerID(peerIp), linkId)
			fmt.Println(err)
		} else if wasDeleted {
			deleted = true
		}
	}
	return
}

func (man *RemoteManager) getPossibilities(linkId graph.LinkID) ([]string, []getResponse) {
	peerIps := []string{}
	possibilities := []getResponse{}
	for _, peerIp := range []string{"192.168.33.10:8080", "192.168.33.11:8080", "10.0.2.15:8080", ""} {

		if man.peerId == peerID(peerIp) {
			//do not connect to self
			continue
		}

		response, err := man.requestState(peerID(peerIp), linkId)
		if err != nil {
			man.unableToSync(peerID(peerIp), linkId)
			continue
		}

		peerIps = append(peerIps, peerIp)
		possibilities = append(possibilities, response)
	}

	return peerIps, possibilities
}
