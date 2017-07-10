package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/remote/peer"
	"fmt"
	"net"
	"os"
)

const DNS_ENTRY = "%s.%s.svc.cluster.local"

var SERVICE = os.Getenv("SERVICE")
var NAMESPACE = os.Getenv("NAMESPACE")

func LookupPeerIps() ([]peer.PeerID, error) {
	ips, err := net.LookupIP(fmt.Sprintf(DNS_ENTRY, SERVICE, NAMESPACE))
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
