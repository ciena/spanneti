package resolver

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"reflect"
	"runtime"
)

func GetPhysicalInterface() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Println("----- Local -----")

	{
		ownPid, err := netns.New()
		if err != nil {
			return err
		}
		defer ownPid.Close()

		ownHandle, err := netlink.NewHandleAt(ownPid)
		if err != nil {
			return err
		}

		links, err := ownHandle.LinkList()
		if err != nil {
			return err
		}

		for num, link := range links {
			fmt.Println(num, link)

			switch link := link.(type) {
			case *netlink.Bond:
				fmt.Println(link.Name)
			case *netlink.Bridge:
				fmt.Println(link.Name)
			case *netlink.Device:
				fmt.Println(link.Name)
			case *netlink.Dummy:
				fmt.Println(link.Name)
			case *netlink.GenericLink:
				fmt.Println(link.Name)
			case *netlink.Gretap:
				fmt.Println(link.Name)
			case *netlink.Iptun:
				fmt.Println(link.Name)
			case *netlink.Veth:
				fmt.Println(link.Name)
			case *netlink.Vti:
				fmt.Println(link.Name)
			default:
				fmt.Println("Unknown type:", reflect.TypeOf(link))
			}
		}
	}

	fmt.Println("----- Host -----")

	{
		hostPid, err := netns.GetFromPid(1)
		if err != nil {
			return err
		}
		defer hostPid.Close()

		hostHandle, err := netlink.NewHandleAt(hostPid)
		if err != nil {
			return err
		}

		links, err := hostHandle.LinkList()
		if err != nil {
			return err
		}

		for num, link := range links {
			fmt.Println(num, link)

			switch link := link.(type) {
			case *netlink.Bond:
				fmt.Println(link.Name)
			case *netlink.Bridge:
				fmt.Println(link.Name)
			case *netlink.Device:
				fmt.Println(link.Name)
			case *netlink.Dummy:
				fmt.Println(link.Name)
			case *netlink.GenericLink:
				fmt.Println(link.Name)
			case *netlink.Gretap:
				fmt.Println(link.Name)
			case *netlink.Iptun:
				fmt.Println(link.Name)
			case *netlink.Veth:
				fmt.Println(link.Name)
			case *netlink.Vti:
				fmt.Println(link.Name)
			default:
				fmt.Println("Unknown type:", reflect.TypeOf(link))
			}
		}
	}

	return nil
}
