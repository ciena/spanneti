package resolver

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func DeleteContainerRemoteInterface(ethName string, containerPid int) (bool, int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get container handle
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return false, 0, err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return false, 0, err
	}

	//if the interface exists
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		//if the interface type is veth
		if link, isVxlan := link.(*netlink.Vxlan); isVxlan {
			//delete
			err := containerHandle.LinkDel(link)
			return true, link.VxlanId, err
		}
	}
	return false, 0, nil
}
