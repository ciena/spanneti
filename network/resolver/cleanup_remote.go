package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func TeardownRemoteContainerLink(ethName string, containerPid int) error {
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
		if _, isVxlan := link.(*netlink.Vxlan); isVxlan {
			//if an interface of the proper type already exists, there's nothing to do
			if err := containerHandle.LinkDel(link); err != nil {
				return err
			}
			fmt.Println("Interface", ethName, "deleted")
			return nil
		}
	}

	return nil
}
