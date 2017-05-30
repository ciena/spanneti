package resolver

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func DeleteContainerInterface(ethName string, containerPid int) (bool, error) {
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
		//delete
		err := containerHandle.LinkDel(link)
		return true, err
	}
	return false, nil
}
