package ip

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"time"
)

const IP_ROUTER_DNS_ENTRY = "ip-router.default.svc.cluster.local"

//requestDelete DELETES a link
func (p tenantIpPlugin) requestSetup(ip TenantIP) error {
	return request(ip, p.fabricIp, http.MethodPost)
}

//requestDelete DELETES a link
func (p tenantIpPlugin) requestDelete(ip TenantIP) error {
	return request(ip, p.fabricIp, http.MethodDelete)
}

type entry struct {
	Prefix  TenantIP `json:"prefix"`
	NextHop string   `json:"nextHop"`
}

func request(ip TenantIP, fabricIp, method string) error {
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 1 * time.Second,
			}).Dial,
		},
	}

	ips, err := net.LookupIP(IP_ROUTER_DNS_ENTRY)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.New("No IPs found for " + IP_ROUTER_DNS_ENTRY)
	}

	data, err := json.Marshal(entry{Prefix: ip, NextHop: fabricIp})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		method,
		"http://"+ips[0].String()+":8080/public-ip",
		bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("Unexpected return code:" + resp.Status)
	}
	return nil
}
