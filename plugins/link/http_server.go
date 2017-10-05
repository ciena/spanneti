package link

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func (man *linkPlugin) newHttpHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/resync", man.resyncHandler).Methods(http.MethodPost)
	//r.HandleFunc("/peer/{peerId}/link/", man.listLinksHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{fabricIp}/link/{linkId}", man.getLinkHandler).Methods(http.MethodGet)
	r.HandleFunc("/peer/{fabricIp}/link/{linkId}", man.updateLinkHandler).Methods(http.MethodPut)
	r.HandleFunc("/peer/{fabricIp}/link/{linkId}", man.deleteLinkHandler).Methods(http.MethodDelete)
	return r
}

func (man *linkPlugin) resyncHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(http.StatusBadRequest)
		return
	}

	var linkIds []linkID
	err = json.Unmarshal(data, &linkIds)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//spin off process to run the events
	go func() {
		for _, linkId := range linkIds {
			man.event("link", linkId)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

type getResponse struct {
	TunnelId tunnelID `json:"tunnel-id"`
	FabricIp string   `json:"fabric-ip"`
	Setup    bool     `json:"setup"`
}

func (man *linkPlugin) getLinkHandler(w http.ResponseWriter, r *http.Request) {
	fabricIp := mux.Vars(r)["fabricIp"]
	linkId := linkID(mux.Vars(r)["linkId"])

	if related := man.GetRelatedTo(PLUGIN_NAME, "link", linkId).([]LinkData); len(related) != 1 {
		//linkId not found, or not available
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := getResponse{FabricIp: man.fabricIp}

	//if this side already has a tunnel set up
	if tunnelId, allocated := man.tunnelMan.tunnelFor(fabricIp, linkId); allocated {
		//return existing
		response.TunnelId = tunnelId
		response.Setup = true
	} else {
		//if undefined, propose lowest available tunnelId
		response.TunnelId = *man.tunnelMan.firstAvailableTunnelId()
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

func (man *linkPlugin) updateLinkHandler(w http.ResponseWriter, r *http.Request) {
	fabricIp := mux.Vars(r)["fabricIp"]
	linkId := linkID(mux.Vars(r)["linkId"])

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

	related := man.GetRelatedTo(PLUGIN_NAME, "link", linkId).([]LinkData)
	if len(related) != 1 {
		//linkId not found, or not available
		w.WriteHeader(http.StatusNotFound)
		return
	}

	containerPid, err := man.GetContainerPid(related[0].ContainerID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allocated, err := man.tunnelMan.allocate(linkId, related[0].GetIfaceFor(linkId), containerPid, proposal.TunnelId, fabricIp)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if allocated {
		proposal.status = http.StatusCreated
	} else {
		//if already exists, and has a higher ID, recommend existing
		if tunnelId, exists := man.tunnelMan.tunnelFor(fabricIp, linkId); exists && tunnelId > proposalRequest.TunnelId {
			proposal.TryTunnelId = &tunnelId
		} else {
			//otherwise use original recommendation
			proposal.TryTunnelId = man.tunnelMan.nextAvailableTunnelId(proposalRequest.TunnelId)
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

func (man *linkPlugin) deleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	fabricIp := mux.Vars(r)["fabricIp"]
	linkId := linkID(mux.Vars(r)["linkId"])

	if _, exists := man.tunnelMan.tunnelFor(fabricIp, linkId); exists {
		defer func() {
			//send an event for this linkId
			go func() {
				man.event("link", linkId)
			}()
		}()
		if err := man.tunnelMan.deallocate(linkId); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
