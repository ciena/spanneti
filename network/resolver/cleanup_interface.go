package resolver

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func DeleteContainerPeerInterface(ethName string, containerPid int) (bool, error) {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get the namespace handles for each container
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return false, err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return false, err
	}

	//if the interface exists
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		//if the interface type is veth
		if _, isVeth := link.(*netlink.Veth); isVeth {
			//delete
			err := containerHandle.LinkDel(link)
			return true, err
		}
	}
	return false, nil
}

func DeleteContainerOltInterface(ethName string, containerPid int) (bool, error) {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get the namespace handles for each container
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return false, err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return false, err
	}

	//if the interface exists
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		//if the interface type is vlan
		if _, isVlan := link.(*netlink.Vlan); isVlan {
			//delete
			err := containerHandle.LinkDel(link)
			return true, err
		}
	}
	return false, nil
}
