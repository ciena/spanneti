package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/remote/peer"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"time"
)

func (man *RemoteManager) runServer() {
	r := mux.NewRouter()
	r.HandleFunc("/resync", man.resyncHandler).Methods(http.MethodPost)
	//r.HandleFunc("/peer/{peerId}/link/", man.listLinksHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{fabricIp}/link/{linkId}", man.getLinkHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{fabricIp}/link/{linkId}", man.updateLinkHandler).Methods(http.MethodPut)
	r.HandleFunc("/peer/{fabricIp}/link/{linkId}", man.deleteLinkHandler).Methods(http.MethodDelete)

	srv := &http.Server{
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		Handler:      r,
		Addr:         string(man.peerId) + ":8080",
	}
	srv.ListenAndServe()
}

type linkResponse struct {
	LinkId   graph.LinkID   `json:"link-id"`
	TunnelId *peer.TunnelID `json:"tunnel-id,omitempty"`
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

//func (man *RemoteManager) listLinksHandler(w http.ResponseWriter, r *http.Request) {
//	peer, err := man.getPeer(peerID(mux.Vars(r)["peerId"]))
//	if err != nil {
//		fmt.Println(err)
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	peer.mutex.Lock()
//	defer peer.mutex.Unlock()
//
//	//stream out the list
//	first := true
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte{'['})
//	for linkId, tunnelId := range peer.tunnelFor {
//		data, err := json.Marshal(linkResponse{
//			LinkId:   linkId,
//			TunnelId: &tunnelId,
//		})
//		if err != nil {
//			fmt.Println(err)
//		}
//
//		//add commas to list
//		if first {
//			first = false
//		} else {
//			w.Write([]byte{','})
//		}
//		w.Write(data)
//	}
//	w.Write([]byte{']'})
//}

type getResponse struct {
	TunnelId peer.TunnelID `json:"tunnel-id"`
	FabricIp string        `json:"fabric-ip"`
	Setup    bool          `json:"setup"`
}

func (man *RemoteManager) getLinkHandler(w http.ResponseWriter, r *http.Request) {
	fabricIp := mux.Vars(r)["fabricIp"]
	linkId := graph.LinkID(mux.Vars(r)["linkId"])

	if related := man.graph.GetRelatedTo(linkId); len(related) != 1 {
		//linkId not found, or not available
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := getResponse{FabricIp: man.fabricIp}

	//if this side already has a tunnel set up
	if tunnelId, allocated := man.tunnelMan.TunnelFor(fabricIp, linkId); allocated {
		//return existing
		response.TunnelId = tunnelId
		response.Setup = true
	} else {
		//if undefined, propose lowest available tunnelId
		response.TunnelId = *man.tunnelMan.FirstAvailableTunnelId()
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
	TunnelId peer.TunnelID `json:"tunnel-id"`
	//if a tunnelId is given, we must respond with the same tunnelId (signalling setup complete), or an unused tunnelId
	//if a tunnelId is not given, we must respond with the current tunnelId (if defined), else an empty tunnelId
}

type linkProposalResponse struct {
	linkProposal
	TryTunnelId *peer.TunnelID `json:"try-tunnel-id,omitempty"`
	status      int            `json:"-"`
}

func (man *RemoteManager) updateLinkHandler(w http.ResponseWriter, r *http.Request) {
	fabricIp := mux.Vars(r)["fabricIp"]
	linkId := graph.LinkID(mux.Vars(r)["linkId"])

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

	containerPid, err := man.getContainerPid(related[0].ContainerId)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allocated, err := man.tunnelMan.Allocate(linkId, related[0].GetIfaceFor(linkId), containerPid, proposal.TunnelId, fabricIp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if allocated {
		proposal.status = http.StatusCreated
	} else {
		//if already exists, and has a higher ID, recommend existing
		if tunnelId, exists := man.tunnelMan.TunnelFor(fabricIp, linkId); exists && tunnelId > proposalRequest.TunnelId {
			proposal.TryTunnelId = &tunnelId
		} else {
			//otherwise use original recommendation
			proposal.TryTunnelId = man.tunnelMan.NextAvailableTunnelId(proposalRequest.TunnelId)
		}
		proposal.status = http.StatusConflict
	}

	//return
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
	fabricIp := mux.Vars(r)["fabricIp"]
	linkId := graph.LinkID(mux.Vars(r)["linkId"])

	if _, exists := man.tunnelMan.TunnelFor(fabricIp, linkId); exists {
		defer func() {
			//send an event for this linkId
			go func() {
				man.eventBus <- linkId
			}()
		}()
		if err := man.tunnelMan.Deallocate(linkId); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
