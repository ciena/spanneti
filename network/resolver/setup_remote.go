package resolver

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const FABRIC_INTERFACE_NAME = "fabric"

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

	fabricLink, err := hostHandle.LinkByName(FABRIC_INTERFACE_NAME)
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

func SetupRemoteContainerLink(ethName string, containerPid int, tunnelId int, peerFabricIp string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Println("Dummy setup", ethName, "to", peerFabricIp, "via", tunnelId)

	name, err := fabricInterfaceName(peerFabricIp, tunnelId)
	if err != nil {
		return err
	}

	fmt.Println(name)

	//get container handle
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return err
	}

	fmt.Println("Cleanup")

	//clean up previous
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		if _, isVxlan := link.(*netlink.Vxlan); isVxlan {
			//if an interface of the proper type already exists, there's nothing to do
			fmt.Println("Interface", ethName, "already exists")
			return nil
		} else {
			//if the container exists, but is not a vxlan link, delete it
			fmt.Println("deleting existing", ethName)
			if err := containerHandle.LinkDel(link); err != nil {
				return err
			}
		}
	}

	//get host handle
	hostNs, err := netns.GetFromPid(1)
	if err != nil {
		return err
	}
	defer hostNs.Close()
	hostHandle, err := netlink.NewHandleAt(hostNs)
	if err != nil {
		return err
	}

	fmt.Println("Cleanup #2")

	//delete any pre-existing devices
	if link, err := hostHandle.LinkByName("cord-vxlan-link"); err == nil {
		if err := hostHandle.LinkDel(link); err != nil {
			return err
		}
	}

	//ip link add f0A060104000001 type vxlan id 1 remote 10.6.1.4 dev fabric
	fabricLink, err := hostHandle.LinkByName(FABRIC_INTERFACE_NAME)
	if err != nil {
		return err
	}

	fmt.Println("Create")

	fmt.Println(tunnelId, peerFabricIp, fabricLink.Attrs().Index)

	//ip link add vxlan0 type vxlan id 0 group 10.6.2.3 dev fabric dstport 4789

	//create veth pair
	link := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name: "cord-vxlan-link",
			//TODO: MTU: ???
		},
		VxlanId:      tunnelId,
		Port:         4789,
		Group:        net.ParseIP(peerFabricIp),
		VtepDevIndex: fabricLink.Attrs().Index,
		//what is required for a pure point-to-point L2 network?
		Learning: true,
		L2miss:   true,
		L3miss:   false,
	}
	if err := hostHandle.LinkAdd(link); err != nil {
		return err
	}

	fmt.Println("Move NS")

	//push interface into container
	if err := moveNsUnsafe(link, ethName, containerPid, hostHandle, containerHandle); err != nil {
		return err
	}
	//get created devices
	//inject into container

	return nil
}

func TeardownRemoteContainerLink(linkId graph.LinkID) error {
	fmt.Println("Dummy teardown", linkId)
	return nil
}

func fabricInterfaceName(ip string, tunnelId int) (string, error) {
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
