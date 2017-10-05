package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"os"
	"runtime"
	"strconv"
)

func SetupContainerOLTInterface(ethName string, containerPid int, sTag, cTag uint16) error {
	_, err := execSelf("setup-container-olt-interface",
		"--eth-name="+ethName,
		"--container-pid="+strconv.Itoa(containerPid),
		"--s-tag="+strconv.Itoa(int(sTag)),
		"--c-tag="+strconv.Itoa(int(cTag)))
	return err
}

func setupContainerOLTInterface(ethName string, containerPid, sTag, cTag int) error {
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

	//get interface names
	OUTER_NAME := fmt.Sprintf("%s.%d", HOST_INTERFACE_NAME, sTag)
	INNER_NAME := fmt.Sprintf("%s.%d.%d", HOST_INTERFACE_NAME, sTag, cTag)

	//get fabric interface
	fabricLink, err := hostHandle.LinkByName(HOST_INTERFACE_NAME)
	if err != nil {
		return err
	}

	//create or get fabric.SSS interface
	outerLink, err := setupVlanAndInjectUnsafe(hostHandle, hostHandle, -1, OUTER_NAME, OUTER_NAME, sTag, fabricLink)
	if err != nil {
		return err
	}

	//create and inject final OLT interface
	innerLink, err := setupVlanAndInjectUnsafe(hostHandle, containerHandle, containerPid, INNER_NAME, ethName, cTag, outerLink)
	if err != nil {
		return err
	}

	//address to add
	addr, err := netlink.ParseAddr("192.168.0.1/24")
	if err != nil {
		return err
	}

	//do not add if already exists
	addrs, err := containerHandle.AddrList(innerLink, 0)
	if err != nil {
		return err
	}
	for _, address := range addrs {
		if address.IP.Equal(addr.IP) {
			fmt.Fprintln(os.Stderr, "Address", addr.IP, "already assigned")
			return nil
		}
	}

	//add address
	if err := containerHandle.AddrAdd(innerLink, addr); err != nil {
		return err
	}
	return nil
}

//setupVlanAndInjectUnsafe creates a vlan interface across namespaces (if an appropriate on doesn't already exist)
func setupVlanAndInjectUnsafe(workingHandle, destHandle *netlink.Handle, destPid int, tempName, ethName string, vlanId int, parent netlink.Link) (netlink.Link, error) {
	isCrossNamespace := workingHandle != destHandle

	if link, err := tryRecoverVlanUnsafe(destHandle, ethName, parent); err != nil {
		return nil, err
	} else if link != nil {
		return link, nil
	}

	if isCrossNamespace {
		//delete any pre-existing devices
		if link, err := workingHandle.LinkByName(tempName); err == nil {
			if err := workingHandle.LinkDel(link); err != nil {
				return nil, err
			}
		}
	}

	//create vlan interface
	link := &netlink.Vlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        tempName,
			ParentIndex: parent.Attrs().Index,
		},
		VlanId: vlanId,
	}
	if err := workingHandle.LinkAdd(link); err != nil {
		return nil, err
	}

	if isCrossNamespace {
		//push interface into container
		if err := moveNsUnsafe(link, ethName, destPid, workingHandle, destHandle); err != nil {
			return nil, err
		}
	} else {
		//bring up the interface
		if err := workingHandle.LinkSetUp(link); err != nil {
			return nil, err
		}
	}
	fmt.Fprintln(os.Stderr, "Setup", ethName, "OK")

	return link, nil
}

//tryRecoverExistingUnsafe attempts to recover the given interface, and returns it if found
func tryRecoverVlanUnsafe(handle *netlink.Handle, ethName string, parent netlink.Link) (netlink.Link, error) {
	//check for existing interfaces
	if link, err := handle.LinkByName(ethName); err == nil {
		//ensure correct type (vlan) and parent
		if _, isVlan := link.(*netlink.Vlan); isVlan && link.Attrs().ParentIndex == parent.Attrs().Index {
			//if the interface is set up correctly
			fmt.Fprintln(os.Stderr, "Interface", ethName, "already exists")
			//ensure the interface is up
			if link.Attrs().OperState != netlink.OperUp {
				if err := handle.LinkSetUp(link); err != nil {
					return nil, err
				}
			}
			//already set up!
			return link, nil
		} else {
			//if the link is not correctly configured, delete it
			if err := handle.LinkDel(link); err != nil {
				return nil, err
			}
		}
	}
	//interface not found
	return nil, nil
}
