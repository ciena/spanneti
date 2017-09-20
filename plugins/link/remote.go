package link

import (
	"errors"
	"fmt"
	"net"
	"time"
)

func determineOwnId() peerID {
	var ownId peerID
	backoff := 1
	for ownId == "" {
		fmt.Print("Determining own IP... ")
		//get peers' IPs
		peerIps, err := LookupPeerIps()
		if err != nil {
			fmt.Printf("Error, will retry (%ds)\n", backoff)
			fmt.Println(err)
		} else {
			//get interfaces
			ifaces, err := net.InterfaceAddrs()
			if err != nil {
				fmt.Printf("Error, will retry (%ds)\n", backoff)
				fmt.Println(err)
			} else {
				//shared IPv4 address is ours
				for _, iface := range ifaces {
					if ipnet, ok := iface.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
						if ipnet.IP.To4() != nil {
							for _, peerIp := range peerIps {
								if peerIp == peerID(ipnet.IP.String()) {
									ownId = peerIp
								}
							}
						}
					}
				}

				if ownId == "" {
					fmt.Printf("Unknown, will retry (%ds)\n", backoff)
				}
			}
		}

		if ownId == "" {
			time.Sleep(time.Second * time.Duration(backoff))
			backoff *= 2
			if backoff > 8 {
				backoff = 8
			}
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
func (man *linkPlugin) tryConnect(linkId linkID, ethName string, containerPid int) (bool, error) {
	peerIps, possibilities := man.getPeersWithLink(linkId)

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
		//until setup remotely
		for !setup {
			//try to setup remotely
			var err error
			setup, tunnelId, err = man.requestSetup(peerId, linkId, tunnelId)
			if err != nil {
				man.resync(peerId, linkId)
				return false, err
			}

			//if not setup remotely, verify that the suggested tunnelId is valid
			if !setup {
				if tunnel := man.tunnelMan.thisOrNextAvailableTunnelId(tunnelId); tunnel == nil {
					man.resync(peerId, linkId)
					return false, errors.New("Out of tunnelIds?")
				} else {
					tunnelId = *tunnel
				}
			}
		}

		//now that it's setup remotely, try to setup locally
		allocated, err := man.tunnelMan.allocate(linkId, ethName, containerPid, tunnelId, fabricIp)
		if err != nil {
			man.resync(peerId, linkId)
			return false, err
		}

		if allocated {
			//setup complete!
			localSetup = true
		} else {
			//go to next available tunnel ID
			tunnel := man.tunnelMan.nextAvailableTunnelId(tunnelId)
			//if already exists, and has a higher tunnelId, recommend existing
			if existingTunnelId, exists := man.tunnelMan.tunnelFor(fabricIp, linkId); exists && existingTunnelId > tunnelId {
				tunnel = &existingTunnelId
			}

			if tunnel == nil {
				man.resync(peerId, linkId)
				return false, errors.New("Out of tunnelIds?")
			} else {
				tunnelId = *tunnel
			}

			if err := man.requestDelete(peerId, linkId); err != nil {
				man.resync(peerId, linkId)
				return false, err
			}
			setup = false
		}
	}
	return true, nil
}

func (man *linkPlugin) tryCleanup(linkId linkID) error {
	peers, err := LookupPeerIps()
	if err != nil {
		panic(err)
	}

	err = man.tunnelMan.deallocate(linkId)

	for _, peerId := range peers {
		if man.peerId == peerId {
			//do not connect to self
			continue
		}

		if err := man.requestDelete(peerId, linkId); err != nil {
			man.resync(peerId, linkId)
			fmt.Println(err)
			continue
		}
	}
	return err
}

func (man *linkPlugin) getPeersWithLink(linkId linkID) ([]peerID, []getResponse) {
	var peerIps []peerID
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
			man.resync(peerId, linkId)
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
