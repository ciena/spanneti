// +build darwin

package resolver

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"os"
)

var HOST_INTERFACE_NAME = os.Getenv("HOST_INTERFACE_NAME")

var fabricIp string

func GetFabricIp() (string, error) {
	if fabricIp != "" {
		return fabricIp, nil
	}

	fmt.Print("Determining fabric IP... ")
	data, err := execSelf("get-fabric-ip")
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, &fabricIp); err != nil {
		return "", err
	}
	fmt.Println(fabricIp)
	return fabricIp, nil
}

func getFabricIp() (string, error) {
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
