package resolver

import "github.com/vishvananda/netlink"

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
