package resolver

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"regexp"
	"runtime"
	"strconv"
)

func GetSharedOLTInterfaces() []uint16 {
	output, err := execSelf("get-shared-olt-interfaces")
	if err != nil {
		panic(err)
	}
	var resp []uint16
	if err := json.Unmarshal(output, &resp); err != nil {
		panic(err)
	}
	return resp
}

func getSharedOLTInterfaces() ([]uint16, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get host handle
	hostNs, err := netns.GetFromPid(1)
	if err != nil {
		return []uint16{}, errors.Wrap(err, "unable to get host namespace")
	}
	defer hostNs.Close()
	hostHandle, err := netlink.NewHandleAt(hostNs)
	if err != nil {
		return []uint16{}, errors.Wrap(err, "unable to get host handle")
	}

	//get interface names
	regex := regexp.MustCompile(`^` + HOST_INTERFACE_NAME + `\.(\d+)$`)

	links, err := hostHandle.LinkList()
	if err != nil {
		return []uint16{}, errors.Wrap(err, "unable to list links")
	}

	sTags := make([]uint16, 0)
	for _, link := range links {
		if subMatches := regex.FindStringSubmatch(link.Attrs().Name); len(subMatches) > 1 {
			id, err := strconv.Atoi(subMatches[1])
			if err != nil {
				return []uint16{}, errors.Wrapf(err, "unable to parse int from '%s'", subMatches[1])
			}
			sTags = append(sTags, uint16(id))
		}
	}
	return sTags, nil
}

func deleteContainerOltInterface(ethName string, containerPid int) (bool, error) {
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

func DeleteSharedOLTInterface(sTag uint16) error {
	_, err := execSelf("delete-shared-olt-interface",
		"--s-tag="+strconv.Itoa(int(sTag)))
	return err
}

//DeleteSharedOltInterface remove the interface named "fabric.<s-tag>"
func deleteSharedOLTInterface(sTag int) error {
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
	SHARED_NAME := fmt.Sprintf("%s.%d", HOST_INTERFACE_NAME, sTag)

	//check for existing interfaces
	if link, err := hostHandle.LinkByName(SHARED_NAME); err == nil {
		if err := hostHandle.LinkDel(link); err != nil {
			return err
		}
		fmt.Printf("Removed OLT interface: fabric.%d\n", sTag)
	}

	return nil
}
