package remote

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/khagerma/cord-networking/network/graph"
	"io/ioutil"
	"net/http"
)

type linkResponse struct {
	LinkId   graph.LinkID `json:"link-id"`
	TunnelId *tunnelID    `json:"tunnel-id,omitempty"`
}

func (man *RemoteManager) listLinksHandler(w http.ResponseWriter, r *http.Request) {
	peer := man.getPeer(peerID(mux.Vars(r)["peerId"]))
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

	peer := man.getPeer(peerID(mux.Vars(r)["peerId"]))
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

	if related := man.graph.GetRelatedTo(linkId); len(related) != 1 {
		//linkId not found, or not available
		w.WriteHeader(http.StatusNotFound)
		return
	}

	peer := man.getPeer(peerID(mux.Vars(r)["peerId"]))
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
		//if unallocated, allocate
		peer.allocate(linkId, proposalRequest.TunnelId)
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
	peer := man.getPeer(peerID(mux.Vars(r)["peerId"]))
	peer.mutex.Lock()
	defer peer.mutex.Unlock()
}

type peerLinkID struct {
	peerId peerID
	linkId graph.LinkID
}

//list //startup action; list all so differences can be determined, and new creates/deletes sent
////slice of linkResponses
//
//put //propose new links (abnormal response: [409: conflict] tunnelId proposal)
////linkResponse
//
//delete //remove an existing link
////no content
//
//get //get link if it exists (unused)
//post //force update link (delete old and re-create) (unused)

type listResponse struct {
}

type deleteRequest struct {
	linkId graph.LinkID `json:"link-id"`
}

//how to preserve existing traffic?

//if the link

func linkHandler(http.ResponseWriter, *http.Request) {
	//responses: SETUP_COMPLETE, NO_MATCHING_LINK, PROPOSE_NEW_TUNNEL_ID

	//if provided id is valid, accept the request
	//if provided id is invalid, propose the next available id

	//if given LinkID does not exist, simply respond with "not found"
	//if t
}
