package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func DeleteContainerRemoteInterface(ethName string, containerPid int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get container handle
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		//ns not longer exists, interface must not exist
		fmt.Println(err)
		return nil
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		//ns not longer exists, interface must not exist
		fmt.Println(err)
		return nil
	}

	//if the interface exists
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		//if the interface type is vxlan
		if _, isVxlan := link.(*netlink.Vxlan); isVxlan {
			//delete
			if err := containerHandle.LinkDel(link); err != nil {
				return err
			}
		}
	}
	return nil
}
