package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func SetupOltContainerLink(ethName string, containerPid int, sTag, cTag uint16) error {
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

	fmt.Println("Cleanup")

	//clean up previous
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		if _, isVlan := link.(*netlink.Vlan); isVlan {
			//if an interface of the proper type already exists, there's nothing to do
			fmt.Println("Interface", ethName, "already exists")
			return nil
		} else {
			//if the container exists, but is not a vxlan link, delete it
			if err := containerHandle.LinkDel(link); err != nil {
				return err
			}
		}
	}

	//ensure the interface for the outer tag exists
	fabricLink, err := setupOuterVlanLinkUnsafe(sTag)
	if err != nil {
		return err
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

	NAME := fmt.Sprintf("fabric.%d.%d", sTag, cTag)

	//delete any pre-existing devices
	if link, err := hostHandle.LinkByName(NAME); err == nil {
		hostHandle.LinkDel(link)
	}

	//create vlan interface
	link := &netlink.Vlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        NAME,
			ParentIndex: fabricLink.Attrs().Index,
		},
		VlanId: int(cTag),
	}
	if err := hostHandle.LinkAdd(link); err != nil {
		return err
	}

	//push interface into container
	if err := moveNsUnsafe(link, ethName, containerPid, hostHandle, containerHandle); err != nil {
		return err
	}
	//get created devices
	//inject into container

	return nil
}

func setupOuterVlanLinkUnsafe(vlanTag uint16) (netlink.Link, error) {
	NAME := fmt.Sprintf("fabric.%d", vlanTag)

	//get own handle
	//ownNs, err := netns.Get()
	//if err != nil {
	//	return err
	//}
	//defer ownNs.Close()
	//ownHandle, err := netlink.NewHandleAt(ownNs)
	//if err != nil {
	//	return err
	//}

	//get host handle
	hostNs, err := netns.GetFromPid(1)
	if err != nil {
		return nil, err
	}
	defer hostNs.Close()
	hostHandle, err := netlink.NewHandleAt(hostNs)
	if err != nil {
		return nil, err
	}

	//if already exists, nothing to do
	if link, err := hostHandle.LinkByName(NAME); err == nil {
		fmt.Println("Interface for the outer vlan already exists")
		if link.Attrs().OperState != netlink.OperUp {
			if err := hostHandle.LinkSetUp(link); err != nil {
				return nil, err
			}
		}
		return link, nil
	}

	//ip link add f0A060104000001 type vxlan id 1 remote 10.6.1.4 dev fabric
	fabricLink, err := hostHandle.LinkByName(FABRIC_INTERFACE_NAME)
	if err != nil {
		return nil, err
	}

	//create veth pair
	link := &netlink.Vlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        NAME,
			ParentIndex: fabricLink.Attrs().Index,
		},
		VlanId: int(vlanTag),
	}
	if err := hostHandle.LinkAdd(link); err != nil {
		return nil, err
	}

	//bring up the interface
	if err := hostHandle.LinkSetUp(link); err != nil {
		return nil, err
	}

	//link, err := hostHandle.LinkByName("cord-vxlan-link")
	//if err != nil {
	//	return nil, nil, err
	//}
	//fmt.Println("Move NS")

	//push interface into container
	//if err := moveNsUnsafe(link, ethName, containerPid, hostHandle, ownHandle); err != nil {
	//	return err
	//}
	//get created devices
	//inject into container

	return link, nil
}
