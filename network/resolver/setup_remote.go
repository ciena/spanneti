package resolver

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"strconv"
	"strings"
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

func SetupRemoteContainerLink(peerFabricIp string, linkId graph.LinkID, tunnelId uint32) error {
	fmt.Println("Dummy setup", linkId, "to", peerFabricIp, "via", tunnelId)

	name, err := fabricInterfaceName(peerFabricIp, tunnelId)
	if err != nil {
		return err
	}

	fmt.Println(name)

	return nil
}

func TeardownRemoteContainerLink(peerFabricIp string, linkId graph.LinkID) error {
	fmt.Println("Dummy teardown", linkId, "to", peerFabricIp)
	return nil
}

func fabricInterfaceName(ip string, tunnelId uint32) (string, error) {
	pieces := strings.Split(ip, ".")
	if len(pieces) < 4 {
		return "", errors.New("Invalid fabric IP: " + ip)
	}

	num0, err := strconv.Atoi(pieces[0])
	if err != nil {
		return "", err
	}
	num1, err := strconv.Atoi(pieces[1])
	if err != nil {
		return "", err
	}
	num2, err := strconv.Atoi(pieces[2])
	if err != nil {
		return "", err
	}
	num3, err := strconv.Atoi(pieces[3])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("f%08X%06X", num0<<24+num1<<16+num2<<8+num3, tunnelId), nil
}

//ip link add f0A060104000001 type vxlan id 1 remote 10.6.1.4 dev fabric
