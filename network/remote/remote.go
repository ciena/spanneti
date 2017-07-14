package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/remote/peer"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"net"
	"sync"
	"time"
)

type RemoteManager struct {
	tunnelMan   peer.TunnelManager
	peerId      peer.PeerID
	fabricIp    string
	client      *client.Client
	resyncMutex sync.Mutex
	graph       *graph.Graph
	eventBus    chan<- graph.LinkID
	outOfSync   map[peer.PeerID]map[graph.LinkID]bool
}

func New(network *graph.Graph, client *client.Client, ch chan<- graph.LinkID, allNetGraphs []graph.ContainerNetwork) (*RemoteManager, error) {
	fmt.Print("Determining fabric IP... ")
	ip, err := resolver.DetermineFabricIp()
	if err != nil {
		return nil, err
	}
	fmt.Println(ip)

	man := &RemoteManager{
		tunnelMan: peer.NewManager(),
		peerId:    determineOwnId(),
		fabricIp:  ip,
		client:    client,
		graph:     network,
		eventBus:  ch,
		outOfSync: make(map[peer.PeerID]map[graph.LinkID]bool),
	}

	//scan for existing remote links
	for _, netGraph := range allNetGraphs {
		containerPid, err := man.getContainerPid(netGraph.ContainerId)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for ethName, linkId := range netGraph.Links {
			man.tunnelMan.FindExisting(linkId, ethName, containerPid)
		}
	}

	go man.runServer()
	return man, nil
}

func determineOwnId() peer.PeerID {
	var ownId peer.PeerID
	for ownId == "" {
		fmt.Print("Determining own IP... ")

		peerIps, err := LookupPeerIps()
		if err != nil {
			panic(err)
		}
		ifaces, err := net.InterfaceAddrs()
		if err != nil {
			panic(err)
		}

		//shared IPv4 address is ours
		for _, iface := range ifaces {
			if ipnet, ok := iface.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					for _, peerIp := range peerIps {
						if peerIp == peer.PeerID(ipnet.IP.String()) {
							ownId = peerIp
						}
					}
				}
			}
		}

		if ownId == "" {
			fmt.Println("Unknown, will retry")
			time.Sleep(time.Second)
		} else {
			fmt.Println(ownId)
		}
	}
	return ownId
}

//in order:
//get availability from remote
//provision on remote
//provision locally
func (man *RemoteManager) TryConnect(linkId graph.LinkID, ethName string, containerPid int) (bool, error) {
	peerIps, possibilities := man.getPossibilities(linkId)

	if len(possibilities) != 1 {
		if len(possibilities) == 0 {
			fmt.Println("No remote containers with linkId", linkId)
		} else {
			fmt.Println("Too many remote containers with linkId", linkId)
		}
		return false, nil
	}

	peerId := peerIps[0]
	response := possibilities[0]
	tunnelId := response.TunnelId

	localSetup := false
	setup := response.Setup
	fabricIp := response.FabricIp

	for !localSetup {
		for !setup {
			//
			// ensure the suggested tunnelId is valid on our side
			//

			var err error
			setup, tunnelId, err = man.requestSetup(peerId, linkId, tunnelId)
			if err != nil {
				man.unableToSync(peerId, linkId)
				return false, err
			}
		}

		//now that it's setup remotely, try to setup locally
		allocated, err := man.tunnelMan.Allocate(linkId, ethName, containerPid, tunnelId, fabricIp)
		if err != nil {
			man.unableToSync(peerId, linkId)
			return false, err
		}

		if allocated {
			//setup complete!
			localSetup = true
		} else {
			//go to next available tunnel ID
			tunnel := man.tunnelMan.NextAvailableTunnelId(tunnelId)
			//if already exists, and has a higher tunnelId, recommend existing
			if existingTunnelId, exists := man.tunnelMan.TunnelFor(fabricIp, linkId); exists && existingTunnelId > tunnelId {
				tunnel = &existingTunnelId
			}

			if tunnel == nil {
				man.unableToSync(peerId, linkId)
				return false, errors.New("Out of tunnelIds?")
			} else {
				tunnelId = *tunnel
			}

			if err := man.requestDelete(peerId, linkId); err != nil {
				man.unableToSync(peerId, linkId)
				return false, err
			}
			setup = false
		}
	}
	return true, nil
}

func (man *RemoteManager) TryCleanup(linkId graph.LinkID) error {
	peers, err := LookupPeerIps()
	if err != nil {
		panic(err)
	}

	for _, peerId := range peers {
		if man.peerId == peerId {
			//do not connect to self
			continue
		}

		if err := man.requestDelete(peerId, linkId); err != nil {
			man.unableToSync(peerId, linkId)
			fmt.Println(err)
			continue
		}
	}

	return man.tunnelMan.Deallocate(linkId)
}

func (man *RemoteManager) getPossibilities(linkId graph.LinkID) ([]peer.PeerID, []getResponse) {
	var peerIps []peer.PeerID
	var possibilities []getResponse

	peers, err := LookupPeerIps()
	if err != nil {
		panic(err)
	}

	for _, peerId := range peers {
		if man.peerId == peerId {
			//do not connect to self
			continue
		}

		response, haveLinkId, err := man.requestState(peerId, linkId)
		if err != nil {
			man.unableToSync(peerId, linkId)
			fmt.Println(err)
			continue
		}

		if haveLinkId {
			peerIps = append(peerIps, peerId)
			possibilities = append(possibilities, response)
		}
	}

	return peerIps, possibilities
}

func (man *RemoteManager) getContainerPid(containerId graph.ContainerID) (int, error) {
	if container, err := man.client.ContainerInspect(context.Background(), string(containerId)); err != nil {
		return 0, err
	} else {
		return container.State.Pid, nil
	}
}
