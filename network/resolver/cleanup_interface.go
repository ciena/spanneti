package resolver

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"strconv"
)

func DeleteContainerPeerInterface(ethName string, containerPid int) error {
	_, err := execSelf("delete-container-peer-interface",
		"--eth-name="+ethName,
		"--container-pid="+strconv.Itoa(containerPid))
	return err
}

func deleteContainerPeerInterface(ethName string, containerPid int) error {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get the namespace handles for each container
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return err
	}

	//if the interface exists
	if link, err := containerHandle.LinkByName(ethName); err == nil {
		//if the interface type is veth
		if _, isVeth := link.(*netlink.Veth); isVeth {
			//delete
			return containerHandle.LinkDel(link)
		}
	}
	return nil
}
