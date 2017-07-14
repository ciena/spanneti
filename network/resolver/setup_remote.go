package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"runtime"
)

func SetupRemoteContainerLink(ethName string, containerPid int, tunnelId int, peerFabricIp string) (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get container handle
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return false, err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return false, err
	}

	//clean up previous
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		//if the container exists, but is not a vxlan link, delete it
		fmt.Println("deleting existing", ethName)
		if err := containerHandle.LinkDel(link); err != nil {
			return false, err
		}
	}

	//get host handle
	hostNs, err := netns.GetFromPid(1)
	if err != nil {
		return false, err
	}
	defer hostNs.Close()
	hostHandle, err := netlink.NewHandleAt(hostNs)
	if err != nil {
		return false, err
	}

	//delete any pre-existing devices
	if link, err := hostHandle.LinkByName("cord-vxlan-link"); err == nil {
		if err := hostHandle.LinkDel(link); err != nil {
			return false, err
		}
	}

	//ip link add f0A060104000001 type vxlan id 1 remote 10.6.1.4 dev fabric
	fabricLink, err := hostHandle.LinkByName(FABRIC_INTERFACE_NAME)
	if err != nil {
		return false, err
	}

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
		if err.Error() == "file exists" {
			//assume vxlanId is in use
			return false, nil
		}
		return false, err
	}

	//push interface into container
	if err := moveNsUnsafe(link, ethName, containerPid, hostHandle, containerHandle); err != nil {
		return false, err
	}
	//get created devices
	//inject into container

	return true, nil
}
