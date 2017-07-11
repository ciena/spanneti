package resolver

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func FindExisting(ethName string, containerPid int) (string, *int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get container handle
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return "", nil, err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return "", nil, err
	}

	if link, err := containerHandle.LinkByName(ethName); err == nil {
		if link, isVxlan := link.(*netlink.Vxlan); isVxlan {
			return link.Group.String(), &link.VxlanId, nil
		}
	}
	return "", nil, nil
}
