package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/remote/peer"
	"net"
	"os"
)

var DNS_NAME = os.Getenv("DNS_NAME")

func LookupPeerIps() ([]peer.PeerID, error) {
	ips, err := net.LookupIP(DNS_NAME)
	if err != nil {
		return []peer.PeerID{}, err
	}

	peers := []peer.PeerID{}
	for _, ip := range ips {
		if ip.To4() != nil {
			peers = append(peers, peer.PeerID(ip.String()))
		}
	}

	return peers, nil
}
