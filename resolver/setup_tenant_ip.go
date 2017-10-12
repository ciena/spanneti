package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"os"
	"runtime"
	"strconv"
)

func SetupTenantIpContainerLink(ethName string, containerPid int, tenantIp string) error {
	_, err := execSelf("setup-tenant-ip-container-link",
		"--eth-name="+ethName,
		"--container-pid="+strconv.Itoa(containerPid),
		"--tenant-ip="+tenantIp)
	return err
}

func setupTenantIpContainerLink(ethName string, containerPid int, tenantIp string) error {
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
		if link, isIPVlan := link.(*netlink.IPVlan); isIPVlan {
			fmt.Fprintln(os.Stderr, "Interface", ethName, "already exists")
			return nil
		} else {
			//if the interface exists, but is not a vxlan link, delete it
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
	if link, err := hostHandle.LinkByName("cord-ipvlan-lnk"); err == nil {
		if err := hostHandle.LinkDel(link); err != nil {
			return err
		}
	}

	//ip link add f0A060104000001 type vxlan id 1 remote 10.6.1.4 dev fabric
	fabricLink, err := hostHandle.LinkByName(HOST_INTERFACE_NAME)
	if err != nil {
		return err
	}

	//create veth
	link := &netlink.IPVlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        "cord-ipvlan-lnk",
			ParentIndex: fabricLink.Attrs().Index,
			//TODO: MTU: ???
			//Namespace: containerPid,
			//change name
			//OperState: netlink.OperUp,
		},
		Mode: netlink.IPVLAN_MODE_L3,
	}
	if err := hostHandle.LinkAdd(link); err != nil {
		return err
	}

	//push interface into container
	if err := moveNsUnsafe(link, ethName, containerPid, hostHandle, containerHandle); err != nil {
		return err
	}

	//add address
	addr, err := netlink.ParseAddr(tenantIp)
	if err != nil {
		return err
	}
	if err := containerHandle.AddrAdd(link, addr); err != nil {
		return err
	}
	return nil
}
