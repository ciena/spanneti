package link

import (
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

//requestState GETs a link
func (man *LinkManager) requestState(peerIp peerID, linkId linkID) (getResponse, bool, error) {
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
		"http://"+fmt.Sprint(peerIp)+":8080/peer/"+url.PathEscape(man.fabricIp)+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		nil)
	if err != nil {
		return getResponse{}, false, err
	}

	resp, err := client.Do(request)
	if err != nil {
		//if fails, just go to next
		return getResponse{}, false, err
	}
	defer resp.Body.Close()

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
func (man *LinkManager) requestSetup(peerIp peerID, linkId linkID, tunnelId tunnelID) (bool, tunnelID, error) {
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
		return false, 0, err
	}

	request, err := http.NewRequest(
		http.MethodPut,
		"http://"+string(peerIp)+":8080/peer/"+url.PathEscape(man.fabricIp)+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		bytes.NewReader(data))
	if err != nil {
		return false, 0, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, 0, errors.New("LinkID does not exist on remote peer, linkup failed.")
	}

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {

		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, 0, err
		}

		var response linkProposalResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return false, 0, err
		}

		if resp.StatusCode == http.StatusCreated {
			// yay!  created!
			return true, tunnelId, nil
		}

		if resp.StatusCode == http.StatusConflict {
			//return the suggested tunnelId
			if response.TryTunnelId == nil {
				return false, 0, errors.New("Status was 409 CONFLICT, but try-tunnel-id was not defined. Out of tunnel IDs?")
			}

			//retry setup with this id
			return false, *response.TryTunnelId, nil
		}
	}

	return false, 0, errors.New("Unexpected status code:" + resp.Status)
}

//requestDelete DELETES a link
func (man *LinkManager) requestDelete(peerId peerID, linkId linkID) error {
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
		"http://"+string(peerId)+":8080/peer/"+url.PathEscape(man.fabricIp)+"/link/"+url.PathEscape(fmt.Sprint(linkId)),
		nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Unexpected return code:" + resp.Status)
	}
	return nil
}

//tryResyncUnsafe tries to have the other side resync the given list of links
func (man *LinkManager) tryResyncUnsafe(peerId peerID, linkIds []linkID) error {
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
		"http://"+fmt.Sprint(peerId)+":8080/resync",
		bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted) {
		return errors.New("Unexpected response code: " + resp.Status)
	}
	return nil
}
