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
	link, err := containerHandle.LinkByName(ethName)
	if err == nil {
		//delete
		if err := containerHandle.LinkDel(link); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
