package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/remote/peer"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

//requestInfo GETs peer info
func (man *RemoteManager) requestInfo(peerIp peer.PeerID) (infoResponse, error) {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	request, err := http.NewRequest(
		http.MethodGet,
		"http://"+fmt.Sprint(peerIp)+":8080/info",
		nil)
	if err != nil {
		return infoResponse{}, err
	}

	resp, err := client.Do(request)
	if err != nil {
		//if fails, just go to next
		return infoResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return infoResponse{}, errors.New("Unexpected status code: " + resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return infoResponse{}, err
	}

	var response infoResponse
	if err := json.Unmarshal(data, &response); err != nil {
		//if fails, just go to next
		return infoResponse{}, err
	}

	return response, nil
}

//requestState GETs a link
func (man *RemoteManager) requestState(peerIp peer.PeerID, linkId graph.LinkID) (getResponse, bool, error) {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	request, err := http.NewRequest(
		http.MethodGet,
		"http://"+fmt.Sprint(peerIp)+":8080/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		nil)
	if err != nil {
		return getResponse{}, false, err
	}

	resp, err := client.Do(request)
	if err != nil {
		//if fails, just go to next
		return getResponse{}, false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return getResponse{}, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return getResponse{}, false, errors.New("Unexpected status code: " + resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return getResponse{}, false, err
	}

	var response getResponse
	if err := json.Unmarshal(data, &response); err != nil {
		//if fails, just go to next
		return getResponse{}, false, err
	}

	return response, true, nil
}

//requestSetup PUTs a link
func (man *RemoteManager) requestSetup(peerIp peer.PeerID, linkId graph.LinkID, tunnelId peer.TunnelID) (bool, peer.TunnelID, string, error) {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	data, err := json.Marshal(&linkProposal{TunnelId: tunnelId, FabricIp: man.fabricIp})
	if err != nil {
		return false, tunnelId, "", err
	}

	request, err := http.NewRequest(
		http.MethodPut,
		"http://"+string(peerIp)+":8080/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		bytes.NewReader(data))
	if err != nil {
		return false, tunnelId, "", err
	}

	resp, err := client.Do(request)
	if err != nil {
		return false, tunnelId, "", err
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, tunnelId, "", errors.New("LinkID does not exist on remote peer, linkup failed.")
	}

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {

		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, tunnelId, "", err
		}

		var response linkProposalResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return false, tunnelId, "", err
		}

		if resp.StatusCode == http.StatusCreated {
			// yay!  created!
			return true, tunnelId, response.FabricIp, nil
		}

		if resp.StatusCode == http.StatusConflict {
			//return the suggested tunnelId
			if response.TryTunnelId == nil {
				return false, 0, "", errors.New("Status was 409 CONFLICT, but try-tunnel-id was not defined. Out of tunnel IDs?")
			}

			//retry setup with this id
			return false, *response.TryTunnelId, response.FabricIp, nil
		}
	}

	return false, tunnelId, "", errors.New("Unexpected status code:" + resp.Status)
}

//requestDelete DELETES a link
func (man *RemoteManager) requestDelete(peerId peer.PeerID, linkId graph.LinkID) (bool, error) {
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
		"http://"+string(peerId)+":8080/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, errors.New("Unexpected return code:" + resp.Status)
	}
	return true, nil
}

//tryResyncUnsafe tries to have the other side resync the given list of links
func (man *RemoteManager) tryResyncUnsafe(peerId peer.PeerID, linkIds []graph.LinkID) error {
	client := http.Client{
		Timeout: 300 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).Dial,
		},
	}

	data, err := json.Marshal(&linkIds)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		"http://"+fmt.Sprint(peerId)+":8080/peer/"+url.PathEscape(string(man.peerId))+"/resync",
		bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted) {
		return errors.New("Unexpected response code: " + resp.Status)
	}
	return nil
}
