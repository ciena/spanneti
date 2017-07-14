package resolver

import (
	"errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"syscall"
)

const FABRIC_INTERFACE_NAME = "fabric"

func DetermineFabricIp() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hostPid, err := netns.GetFromPid(1)
	if err != nil {
		return "", err
	}
	defer hostPid.Close()

	hostHandle, err := netlink.NewHandleAt(hostPid)
	if err != nil {
		return "", err
	}

	fabricLink, err := hostHandle.LinkByName(FABRIC_INTERFACE_NAME)
	if err != nil {
		return "", err
	}

	addrs, err := hostHandle.AddrList(fabricLink, syscall.AF_INET)
	if err != nil {
		return "", err
	}
	if len(addrs) != 1 {
		if len(addrs) == 0 {
			return "", errors.New("No IPs have been assigned to the fabric interface.")
		}
		return "", errors.New("Multiple IPs have been assigned to the fabric interface.")
	}

	return addrs[0].IP.String(), nil
}

func moveNsUnsafe(iface netlink.Link, ethName string, containerPid int, ownHandle, containerHandle *netlink.Handle) error {
	//move into the container
	if err := ownHandle.LinkSetNsPid(iface, containerPid); err != nil {
		return err
	}
	//change names and bring online
	if err := containerHandle.LinkSetName(iface, ethName); err != nil {
		return err
	}
	if err := containerHandle.LinkSetUp(iface); err != nil {
		return err
	}
	return nil
}
