package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/khagerma/cord-networking/network/graph"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type RemoteManager struct {
	peerId peerID
	peer   map[peerID]*remotePeer
	mutex  sync.Mutex
	graph  *graph.Graph
}

func New(peerId string, graph *graph.Graph) *RemoteManager {
	man := &RemoteManager{
		peerId: peerID(peerId),
		peer:   make(map[peerID]*remotePeer),
		graph:  graph,
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
				return false, err
			}
			localSetup = true
		} else {
			//go to next available tunnel ID
			if tunnel := peer.nextAvailableTunnelId(tunnelId); tunnel == nil {
				return false, errors.New("Out of tunnelIds?")
			} else {
				tunnelId = *tunnel
			}
			peer.mutex.Unlock()

			if _, err := man.requestDelete(peerIp, linkId); err != nil {
				return false, err
			}
			setup = false
		}
	}
	return true, nil
}

func (man *RemoteManager) TryCleanup(linkId graph.LinkID) (deleted bool) {
	deleted = false
	for _, peerIp := range []string{"localhost:8080", "localhost:8081"} {
		peer := man.getPeer(peerID(peerIp))
		peer.mutex.Lock()
		if err := peer.deallocate(linkId); err != nil {
			fmt.Println(err)
		}
		peer.mutex.Unlock()

		if wasDeleted, err := man.requestDelete(peerIp, linkId); err != nil {
			fmt.Println(err)
		} else if wasDeleted {
			deleted = true
		}
	}
	return
}

func (man *RemoteManager) getPossibilities(linkId graph.LinkID) ([]string, []getResponse) {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	peerIps := []string{}
	possibilities := []getResponse{}
	for _, peerIp := range []string{"localhost:8080", "localhost:8081"} {
		request, err := http.NewRequest(
			http.MethodGet,
			"http://"+peerIp+"/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
			nil)
		if err != nil {
			fmt.Println(err)
			continue
		}

		resp, err := client.Do(request)
		if err != nil {
			//if fails, just go to next
			continue
		}

		if resp.StatusCode == http.StatusOK {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				continue
			}

			var response getResponse
			if err := json.Unmarshal(data, &response); err != nil {
				//if fails, just go to next
				fmt.Println(err)
				continue
			}

			peerIps = append(peerIps, peerIp)
			possibilities = append(possibilities, response)
		}
	}

	return peerIps, possibilities
}

func (man *RemoteManager) requestSetup(peerIp string, linkId graph.LinkID, tunnelId tunnelID) (bool, tunnelID, error) {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	data, err := json.Marshal(&linkProposal{TunnelId: tunnelId})
	if err != nil {
		return false, tunnelId, err
	}

	request, err := http.NewRequest(
		http.MethodPut,
		"http://"+peerIp+"/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		bytes.NewReader(data))
	if err != nil {
		return false, tunnelId, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return false, tunnelId, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, tunnelId, errors.New("LinkID does not exist on remote peer, linkup failed.")
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, tunnelId, err
	}
	fmt.Println(resp.Status, string(data))

	var response linkProposalResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return false, tunnelId, err
	}

	if resp.StatusCode == http.StatusCreated {
		// yay!  created!
		return true, tunnelId, nil

	} else if resp.StatusCode == http.StatusConflict {
		if response.TryTunnelId == nil {
			return false, 0, errors.New("Status was 409 CONFLICT, but try-tunnel-id was not defined. Out of tunnel IDs?")
		}

		//setup with id
		return false, *response.TryTunnelId, nil
	}

	return false, tunnelId, errors.New("Unexpected status code:" + resp.Status)
}

func (man *RemoteManager) requestDelete(peerIp string, linkId graph.LinkID) (bool, error) {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	request, err := http.NewRequest(
		http.MethodDelete,
		"http://"+peerIp+"/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		nil)
	if err != nil {
		return false, err
	}

	fmt.Println(request.URL)
	resp, err := client.Do(request)
	if err != nil {
		return false, err
	}

	fmt.Println(resp.Status)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return false, errors.New("Unexpected return code:" + resp.Status)
	}
	return resp.StatusCode == http.StatusOK, nil
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
