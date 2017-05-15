package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

func SetupLocalContainerLink(ethName0 string, containerPid0 int, ethName1 string, containerPid1 int) error {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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
	{
		up0 := false
		up1 := false
		//if the interface exists
		link0, err := handle0.LinkByName(ethName0)
		if err == nil {
			//if the link is not up
			if link0.Attrs().OperState != netlink.OperUp {
				//delete
				handle0.LinkDel(link0)
			} else {
				up0 = true
			}
		}
		//if the interface exists
		link1, err := handle1.LinkByName(ethName1)
		if err == nil {
			//if the link is not up
			if link1.Attrs().OperState != netlink.OperUp {
				//delete
				handle1.LinkDel(link1)
			} else {
				up1 = true
			}
		}

		//if both interfaces are up, then there's nothing to do
		if up0 && up1 {
			fmt.Println("Interfaces already up, nothing to do.")
			return nil
		} else {
			//of ether interface is down or unavailable, ensure everything is cleaned up/deleted
			if up0 {
				handle0.LinkDel(link0)
			}
			if up1 {
				handle1.LinkDel(link1)
			}
		}
	}

	//create a new veth pair
	peer0, peer1, err := createVethPairUnsafe()
	if err != nil {
		return err
	}

	//move interfaces into the container namespaces
	ownHandle, err := netlink.NewHandle()
	if err != nil {
		return err
	}
	if err := ownHandle.LinkSetNsPid(peer0, containerPid0); err != nil {
		return err
	}
	if err := ownHandle.LinkSetNsPid(peer1, containerPid1); err != nil {
		return err
	}

	//change names and bring online
	handle0.LinkSetAlias(peer0, ethName0)
	if err := handle0.LinkSetName(peer0, ethName0); err != nil {
		return err
	}
	if err := handle0.LinkSetUp(peer0); err != nil {
		return err
	}

	if err := handle1.LinkSetName(peer1, ethName1); err != nil {
		return err
	}
	if err := handle1.LinkSetUp(peer1); err != nil {
		return err
	}

	//ensure that no addresses are assigned
	if addrs, err := handle0.AddrList(peer0, 0); err != nil {
		return err
	} else {
		for _, addr := range addrs {
			handle1.AddrDel(peer0, &addr)
			fmt.Println(addr)
		}
	}

	if addrs, err := handle1.AddrList(peer1, 0); err != nil {
		return err
	} else {
		for _, addr := range addrs {
			handle1.AddrDel(peer1, &addr)
			fmt.Println(addr)
		}
	}

	return nil
}

func CreateVethPair() (*netlink.Veth, *netlink.Veth, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	return createVethPairUnsafe()

}
func createVethPairUnsafe() (*netlink.Veth, *netlink.Veth, error) {

	ownHandle, err := netlink.NewHandle()
	if err != nil {
		return nil, nil, err
	}
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

	fmt.Println("----- Created -----")
	fmt.Println(link0)
	fmt.Println(link1)

	return link0.(*netlink.Veth), link1.(*netlink.Veth), nil
}
