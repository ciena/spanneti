package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/khagerma/cord-networking/network/graph"
	"github.com/khagerma/cord-networking/network/resolver"
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

func (man *RemoteManager) TryConnect(linkId graph.LinkID) (bool, error) {

	peerIps, possibilities := man.getPossibilities(linkId)

	if len(possibilities) != 1 {
		fmt.Println("There are not exactly two containers with the same link ID.  Ignoring...")
		return false, nil
	}

	peerIp := peerIps[0]
	response := possibilities[0]
	//
	// ensure the suggested tunnelId is valid on our side
	//

	proposal := &linkProposal{TunnelId: response.TunnelId}

	for setup := response.Setup; !setup; {
		if isSetup, err := man.requestSetup(linkId, proposal, peerIp); err != nil {
			return false, err
		} else {
			setup = isSetup
		}
	}

	//set up locally
	if err := resolver.SetupRemoteContainerLink(peerIp, linkId, uint64(response.TunnelId)); err != nil {
		return false, err
	}
	return true, nil
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

		fmt.Println(request.URL)
		resp, err := client.Do(request)
		if err != nil {
			//if fails, just go to next
			fmt.Println(err)
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
				continue
			}

			peerIps = append(peerIps, peerIp)
			possibilities = append(possibilities, response)
		}
	}

	return peerIps, possibilities
}

func (man *RemoteManager) requestSetup(linkId graph.LinkID, proposal *linkProposal, peerIp string) (bool, error) {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	data, err := json.Marshal(proposal)
	if err != nil {
		return false, err
	}

	request, err := http.NewRequest(
		http.MethodPut,
		"http://"+peerIp+"/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		bytes.NewReader(data))
	if err != nil {
		return false, err
	}

	fmt.Println(request.URL, string(data))
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		//LinkID does not exist on this peer, linkup failed
		return false, nil
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	fmt.Println(resp.Status, string(data))

	var response linkProposalResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusCreated {

		//setup local connection
		fmt.Println("ID to verify:", proposal.TunnelId)

		//done
		return true, nil

	} else if resp.StatusCode == http.StatusConflict {
		if response.TryTunnelId == nil {
			return false, errors.New("Status was 409 CONFLICT, but try-tunnel-id was not defined. Out of tunnel IDs?")
		}

		proposal.TunnelId = *response.TryTunnelId

		//setup with id
		return false, nil
	}

	return false, errors.New("Unexpected status code:" + resp.Status)
}

func (man *RemoteManager) runServer() {
	r := mux.NewRouter()
	r.HandleFunc("/peer/{peerId}/link/", man.listLinksHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{peerId}/link/", man.updateLinksHandler).Methods(http.MethodPut)
	r.HandleFunc("/peer/{peerId}/link/", man.deleteLinksHandler).Methods(http.MethodDelete)
	r.HandleFunc("/peer/{peerId}/link/{linkId}", man.getLinkHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{peerId}/link/{linkId}", man.updateLinkHandler).Methods(http.MethodPut)
	r.HandleFunc("/peer/{peerId}/link/{linkId}", man.deleteLinkHandler).Methods(http.MethodDelete)

	srv := &http.Server{
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		Handler:      r,
		Addr:         "localhost:8080",
	}
	srv.ListenAndServe()
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
