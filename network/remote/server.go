package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"time"
)

func (man *RemoteManager) runServer() {
	r := mux.NewRouter()
	r.HandleFunc("/info", man.infoHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{peerId}/resync", man.resyncHandler).Methods(http.MethodPost)
	r.HandleFunc("/peer/{peerId}/link/", man.listLinksHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{peerId}/link/", man.updateLinksHandler).Methods(http.MethodPut)
	r.HandleFunc("/peer/{peerId}/link/", man.deleteLinksHandler).Methods(http.MethodDelete)
	r.HandleFunc("/peer/{peerId}/link/{linkId}", man.getLinkHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{peerId}/link/{linkId}", man.updateLinkHandler).Methods(http.MethodPut)
	r.HandleFunc("/peer/{peerId}/link/{linkId}", man.deleteLinkHandler).Methods(http.MethodDelete)

	srv := &http.Server{
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		Handler:      r,
		Addr:         string(man.peerId) + ":8080",
	}
	srv.ListenAndServe()
}

type infoResponse struct {
	FabricIp string `json:"fabric-ip"`
}

func (man *RemoteManager) infoHandler(w http.ResponseWriter, r *http.Request) {
	response := infoResponse{
		FabricIp: man.fabricIp,
	}
	if data, err := json.Marshal(response); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

type linkResponse struct {
	LinkId   graph.LinkID `json:"link-id"`
	TunnelId *tunnelID    `json:"tunnel-id,omitempty"`
}

func (man *RemoteManager) resyncHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(http.StatusBadRequest)
		return
	}

	var links []graph.LinkID
	err = json.Unmarshal(data, &links)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//spin off process to run the events
	go func() {
		for _, link := range links {
			man.eventBus <- link
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

func (man *RemoteManager) listLinksHandler(w http.ResponseWriter, r *http.Request) {
	peer, err := man.getPeer(peerID(mux.Vars(r)["peerId"]))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	//stream out the list
	first := true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{'['})
	for linkId, tunnelId := range peer.tunnelFor {
		data, err := json.Marshal(linkResponse{
			LinkId:   linkId,
			TunnelId: &tunnelId,
		})
		if err != nil {
			fmt.Println(err)
		}

		//add commas to list
		if first {
			first = false
		} else {
			w.Write([]byte{','})
		}
		w.Write(data)
	}
	w.Write([]byte{']'})
}

func (man *RemoteManager) updateLinksHandler(w http.ResponseWriter, r *http.Request) {

}
func (man *RemoteManager) deleteLinksHandler(w http.ResponseWriter, r *http.Request) {

}

type getResponse struct {
	TunnelId tunnelID `json:"tunnel-id"`
	Setup    bool     `json:"setup"`
}

func (man *RemoteManager) getLinkHandler(w http.ResponseWriter, r *http.Request) {
	linkId := graph.LinkID(mux.Vars(r)["linkId"])

	if related := man.graph.GetRelatedTo(linkId); len(related) != 1 {
		//linkId not found, or not available
		w.WriteHeader(http.StatusNotFound)
		return
	}

	peer, err := man.getPeer(peerID(mux.Vars(r)["peerId"]))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	response := getResponse{}

	//if this side already has a tunnel set up
	if tunnelId, allocated := peer.tunnelFor[linkId]; allocated {
		//return existing
		response.TunnelId = tunnelId
		response.Setup = true
	} else {
		//if undefined, propose lowest available tunnelId
		response.TunnelId = *peer.nextAvailableTunnelId(0)
		response.Setup = false
	}

	if data, err := json.Marshal(response); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

type linkProposal struct {
	TunnelId tunnelID `json:"tunnel-id"`
	//if a tunnelId is given, we must respond with the same tunnelId (signalling setup complete), or an unused tunnelId
	//if a tunnelId is not given, we must respond with the current tunnelId (if defined), else an empty tunnelId
}

type linkProposalResponse struct {
	linkProposal
	TryTunnelId *tunnelID `json:"try-tunnel-id,omitempty"`
	status      int       `json:"-"`
}

func (man *RemoteManager) updateLinkHandler(w http.ResponseWriter, r *http.Request) {
	linkId := graph.LinkID(graph.LinkID(mux.Vars(r)["linkId"]))

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var proposalRequest linkProposal
	err = json.Unmarshal(data, &proposalRequest)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	proposal := linkProposalResponse{
		linkProposal: proposalRequest,
		status:       http.StatusInternalServerError,
	}

	related := man.graph.GetRelatedTo(linkId)
	if len(related) != 1 {
		//linkId not found, or not available
		w.WriteHeader(http.StatusNotFound)
		return
	}

	peer, err := man.getPeer(peerID(mux.Vars(r)["peerId"]))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	//if this side already has a tunnel set up
	if currentLinkId, allocated := peer.linkFor[proposalRequest.TunnelId]; allocated {
		if currentLinkId == linkId {
			//already setup, nothing to do
			//accept proposal
			proposal.status = http.StatusCreated
		} else {
			//tunnelId in use, return next available tunnelId
			proposal.TryTunnelId = peer.nextAvailableTunnelId(proposalRequest.TunnelId)
			proposal.status = http.StatusConflict
		}
	} else {
		containerPid, err := man.getContainerPid(related[0].ContainerId)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//if unallocated, allocate
		if err := peer.allocate(linkId, related[0].GetIfaceFor(linkId), containerPid, proposalRequest.TunnelId); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//accept proposal
		proposal.status = http.StatusCreated
	}

	if data, err := json.Marshal(proposal); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(proposal.status)
		w.Write(data)
	}
}

func (man *RemoteManager) deleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	peer, err := man.getPeer(peerID(mux.Vars(r)["peerId"]))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	peer.mutex.Lock()
	defer peer.mutex.Unlock()

	if err := peer.deallocate(graph.LinkID(mux.Vars(r)["linkId"])); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
