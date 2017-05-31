package remote

import (
	"bytes"
	"encoding/json"
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
	for _, peerIp := range []string{"localhost:8080", "localhost:8081"} {
		proposal := linkProposal{LinkId: linkId}
		for retrySetup := true; retrySetup; {

			data, err := json.Marshal(&proposal)
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

			client := http.Client{
				Timeout: 300 * time.Millisecond,
				Transport: &http.Transport{
					Dial: (&net.Dialer{
						Timeout: 100 * time.Millisecond,
					}).Dial,
				},
			}
			fmt.Println(request.URL, string(data))
			resp, err := client.Do(request)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if resp.StatusCode == http.StatusNotFound {
				//LinkID does not exist on this peer
				continue
			}

			if resp.StatusCode == http.StatusCreated ||
				resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusConflict {
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return false, err
				}
				fmt.Println(resp.Status, string(data))

				var response linkProposalResponse
				if err := json.Unmarshal(data, &response); err != nil {
					return false, err
				}

				if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusOK {
					proposal.TunnelId = response.TryTunnelId
				} else if resp.StatusCode == http.StatusCreated {
					if proposal.TunnelId == nil {
						//if was unknown before
						//verify ID is ok
						fmt.Println("ID to verify: ", response.TryTunnelId)
						response.TunnelId = response.TryTunnelId
					}
					//requested ID is setup on remote
					//set up locally
					if err := resolver.SetupRemoteContainerLink(peerIp, proposal.LinkId, uint64(*response.TunnelId)); err != nil {
						return false, err
					}
					return true, nil
				}

				proposal.TunnelId = response.TryTunnelId

				//setup with id
				continue
			}
		}
	}

	return false, nil
}

func (man *RemoteManager) runServer() {
	r := mux.NewRouter()
	r.HandleFunc("/peer/{peerId}/link/", man.listLinksHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{peerId}/link/", man.updateLinksHandler).Methods(http.MethodPut)
	r.HandleFunc("/peer/{peerId}/link/", man.deleteLinksHandler).Methods(http.MethodDelete)
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
