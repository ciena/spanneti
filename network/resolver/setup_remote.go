package resolver

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"syscall"
)

func DetermineFabricIp() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hostPid, err := netns.GetFromPid(1)
	if err != nil {
		return "", err
	}
	defer hostPid.Close()

	hostHandle, err := netlink.NewHandleAt(hostPid)
	if err != nil {
		return "", err
	}

	fabricLink, err := hostHandle.LinkByName("fabric")
	if err != nil {
		return "", err
	}

	addrs, err := hostHandle.AddrList(fabricLink, syscall.AF_INET)
	if err != nil {
		return "", err
	}
	if len(addrs) != 1 {
		if len(addrs) == 0 {
			return "", errors.New("No IPs have been assigned to the fabric interface.")
		}
		return "", errors.New("Multiple IPs have been assigned to the fabric interface.")
	}

	return addrs[0].IP.String(), nil
}

func SetupRemoteContainerLink(peerFabricIp string, linkId graph.LinkID, tunnelId uint64) error {
	fmt.Println("Dummy setup", linkId, "to", peerFabricIp, "via", tunnelId)
	return nil
}

func TeardownRemoteContainerLink(peerFabricIp string, linkId graph.LinkID) error {
	fmt.Println("Dummy teardown", linkId, "to", peerFabricIp)
	return nil
}

//ip link add fabric-10.0.2.8-10 type vxlan id 10 remote 10.0.2.8 dev fabric
