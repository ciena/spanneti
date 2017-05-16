package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

//SetupLocalContainerLink ensures a veth pair connects the specified containers
func SetupLocalContainerLink(ethName0 string, containerPid0 int, ethName1 string, containerPid1 int) error {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get current container's handle
	ownHandle, err := netlink.NewHandle()
	if err != nil {
		return err
	}

	//get the namespace handles for each container
	ns0, err := netns.GetFromPid(containerPid0)
	if err != nil {
		return err
	}
	defer ns0.Close()
	handle0, err := netlink.NewHandleAt(ns0)
	if err != nil {
		return err
	}

	ns1, err := netns.GetFromPid(containerPid1)
	if err != nil {
		return err
	}
	defer ns1.Close()
	handle1, err := netlink.NewHandleAt(ns1)
	if err != nil {
		return err
	}

	//ensure that the interfaces don't already exist
	if up, err := interfacesUpOrDelUnsafe(ethName0, handle0, ethName1, handle1); err != nil {
		return err
	} else if up {
		fmt.Println("Interfaces already up, nothing to do.")
		return nil
	}

	//create a new veth pair
	peer0, peer1, err := createVethPairUnsafe(ownHandle)
	if err != nil {
		return err
	}

	//move interfaces into the container namespaces, rename them, and bring them up
	if err := moveNsUnsafe(peer0, ethName0, containerPid0, ownHandle, handle0); err != nil {
		return err
	}
	if err := moveNsUnsafe(peer1, ethName1, containerPid1, ownHandle, handle1); err != nil {
		return err
	}
	return nil
}

func interfacesUpOrDelUnsafe(ethName0 string, handle0 *netlink.Handle, ethName1 string, handle1 *netlink.Handle) (bool, error) {
	link0, err := interfaceUpOrDelUnsafe(ethName0, handle0)
	if err != nil {
		return false, err
	}
	link1, err := interfaceUpOrDelUnsafe(ethName1, handle1)
	if err != nil {
		return false, err
	}

	if link0 != nil && link1 != nil {
		//if both interfaces are up, then there's nothing to do
		return true, nil
	}

	//if either interface is down or unavailable, ensure everything is cleaned up/deleted
	if link0 != nil {
		if err := handle0.LinkDel(*link0); err != nil {
			return false, err
		}
	}
	if link1 != nil {
		if err := handle1.LinkDel(*link1); err != nil {
			return false, err
		}
	}

	return false, nil
}

//interfaceUpOrDel removes the given interface if it's DOWN.  If the interface exists and is UP, it is returned
func interfaceUpOrDelUnsafe(ethName string, handle *netlink.Handle) (*netlink.Link, error) {
	link, err := handle.LinkByName(ethName)
	if err == nil {
		//if the link is not up
		if link.Attrs().OperState != netlink.OperUp {
			//delete
			if err := handle.LinkDel(link); err != nil {
				return nil, err
			}
		} else {
			//interface is OK
			return &link, nil
		}
	}
	return nil, nil
}

func createVethPairUnsafe(ownHandle *netlink.Handle) (*netlink.Veth, *netlink.Veth, error) {
	var BASE_NAME string = "cord-veth-peer"
	name0 := BASE_NAME + "0"
	name1 := BASE_NAME + "1"

	fmt.Println(name0, name1)

	//delete any pre-existing devices
	if link, err := ownHandle.LinkByName(name0); err == nil {
		ownHandle.LinkDel(link)
	}
	if link, err := ownHandle.LinkByName(name1); err == nil {
		ownHandle.LinkDel(link)
	}

	//create veth pair
	if err := ownHandle.LinkAdd(&netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: name0,
		},
		PeerName: name1,
	}); err != nil {
		return nil, nil, err
	}

	//get created devices
	link0, err := ownHandle.LinkByName(name0)
	if err != nil {
		return nil, nil, err
	}
	link1, err := ownHandle.LinkByName(name1)
	if err != nil {
		return nil, nil, err
	}

	//list info
	if addrs, err := ownHandle.AddrList(link0, 0); err != nil {
		return nil, nil, err
	} else {
		for _, addr := range addrs {
			fmt.Println(addr)
		}
	}

	if addrs, err := ownHandle.AddrList(link1, 0); err != nil {
		return nil, nil, err
	} else {
		for _, addr := range addrs {
			fmt.Println(addr)
		}
	}

	return link0.(*netlink.Veth), link1.(*netlink.Veth), nil
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
