package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

//DeleteSharedOltInterface remove the interface named "fabric.<s-tag>"
func DeleteSharedOltInterface(sTag uint16) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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

	//get interface names
	SHARED_NAME := fmt.Sprintf("%s.%d", FABRIC_INTERFACE_NAME, sTag)

	//check for existing interfaces
	if link, err := hostHandle.LinkByName(SHARED_NAME); err == nil {
		if err := hostHandle.LinkDel(link); err != nil {
			return err
		}
		fmt.Printf("Removed OLT interface: fabric.%d\n", sTag)
	}

	return nil
}
