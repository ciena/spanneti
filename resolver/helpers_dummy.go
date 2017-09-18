// +build darwin

package resolver

import (
	"encoding/json"
	"github.com/vishvananda/netlink"
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
	return "10.6.1.1(dummy)", nil
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
