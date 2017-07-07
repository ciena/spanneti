package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"regexp"
	"runtime"
	"strconv"
)

func GetSharedOLTInterfaces() []uint16 {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get host handle
	hostNs, err := netns.GetFromPid(1)
	if err != nil {
		panic(err)
	}
	defer hostNs.Close()
	hostHandle, err := netlink.NewHandleAt(hostNs)
	if err != nil {
		panic(err)
	}

	//get interface names
	regex := regexp.MustCompile(`^` + FABRIC_INTERFACE_NAME + `\.(\d+)$`)

	links, err := hostHandle.LinkList()
	if err != nil {
		panic(err)
	}

	sTags := make([]uint16, 0)
	for _, link := range links {
		if subMatches := regex.FindStringSubmatch(link.Attrs().Name); len(subMatches) > 1 {
			id, err := strconv.Atoi(subMatches[1])
			if err != nil {
				panic(err)
			}
			sTags = append(sTags, uint16(id))
		}
	}
	return sTags
}

//DeleteSharedOltInterface remove the interface named "fabric.<s-tag>"
func DeleteSharedOLTInterface(sTag uint16) error {
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
