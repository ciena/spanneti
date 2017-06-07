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
	"time"
)

//requestState GETs a link
func (man *RemoteManager) requestState(peerIp peerID, linkId graph.LinkID) (getResponse, error) {
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
		"http://"+fmt.Sprint(peerIp)+"/peer/"+url.PathEscape(string(man.peerId))+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		nil)
	if err != nil {
		return getResponse{}, err
	}

	resp, err := client.Do(request)
	if err != nil {
		//if fails, just go to next
		return getResponse{}, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return getResponse{}, errors.New("Unexpected status code: " + resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return getResponse{}, err
	}

	var response getResponse
	if err := json.Unmarshal(data, &response); err != nil {
		//if fails, just go to next
		return getResponse{}, err
	}

	return response, nil
}

//requestSetup PUTs a link
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

		//retry setup with this id
		return false, *response.TryTunnelId, nil
	}

	return false, tunnelId, errors.New("Unexpected status code:" + resp.Status)
}

//requestDelete DELETES a link
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

	fmt.Println(request.Method, request.URL)
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
