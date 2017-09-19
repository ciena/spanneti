package link

import (
	"net"
	"os"
)

var DNS_NAME = os.Getenv("DNS_NAME")

func LookupPeerIps() ([]peerID, error) {
	ips, err := net.LookupIP(DNS_NAME)
	if err != nil {
		return []peerID{}, err
	}

	peers := []peerID{}
	for _, ip := range ips {
		if ip.To4() != nil {
			peers = append(peers, peerID(ip.String()))
		}
	}

	return peers, nil
}
