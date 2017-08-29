package resolver

import (
	"encoding/json"
	"errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"syscall"
	"os"
)

var HOST_INTERFACE_NAME = os.Getenv("HOST_INTERFACE_NAME")

func DetermineFabricIp() (string, error) {
	data, err := execSelf("determine-fabric-ip")
	if err != nil {
		return "", err
	}
	var fabricIp string
	if err := json.Unmarshal(data, &fabricIp); err != nil {
		return "", err
	}
	return fabricIp, nil
}

func determineFabricIp() (string, error) {
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

	fabricLink, err := hostHandle.LinkByName(HOST_INTERFACE_NAME)
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

func moveNsUnsafe(link netlink.Link, ethName string, containerPid int, ownHandle, containerHandle *netlink.Handle) error {
	//move into the container
	if err := ownHandle.LinkSetNsPid(link, containerPid); err != nil {
		return err
	}
	//change names and bring online
	if err := containerHandle.LinkSetName(link, ethName); err != nil {
		return err
	}
	if err := containerHandle.LinkSetUp(link); err != nil {
		return err
	}
	return nil
}
