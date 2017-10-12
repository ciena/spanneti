package resolver

import (
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"runtime"
	"strconv"
)

func SetupRemoteContainerLink(ethName string, containerPid int, tunnelId int, peerFabricIp string) error {
	_, err := execSelf("setup-remote-container-link",
		"--eth-name="+ethName,
		"--container-pid="+strconv.Itoa(containerPid),
		"--tunnel-id="+strconv.Itoa(tunnelId),
		"--peer-fabric-ip="+peerFabricIp)
	return err
}

func setupRemoteContainerLink(ethName string, containerPid int, tunnelId int, peerFabricIp string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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

	//clean up previous
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		//if the interface exists, but is not a vxlan link, delete it
		if vxlanLink, isVxlan := link.(*netlink.Vxlan); isVxlan && vxlanLink.VxlanId == tunnelId && vxlanLink.Group.Equal(net.ParseIP(peerFabricIp)) {
			//nothing to do
			return nil
		} else {
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

	//delete any pre-existing devices
	if link, err := hostHandle.LinkByName("cord-vxlan-link"); err == nil {
		if err := hostHandle.LinkDel(link); err != nil {
			return err
		}
	}

	//ip link add f0A060104000001 type vxlan id 1 remote 10.6.1.4 dev fabric
	fabricLink, err := hostHandle.LinkByName(HOST_INTERFACE_NAME)
	if err != nil {
		return err
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
			return errors.New("tunnel unavailable")
		}
		return err
	}

	//push interface into container
	if err := moveNsUnsafe(link, ethName, containerPid, hostHandle, containerHandle); err != nil {
		return err
	}

	//add address
	addr, err := netlink.ParseAddr("192.168.0.1/24")
	if err != nil {
		return err
	}
	if err := containerHandle.AddrAdd(link, addr); err != nil {
		return err
	}
	return nil
}
