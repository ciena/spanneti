package resolver

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"strconv"
)

type existing struct {
	FabricIp string `json:"fabric-ip"`
	TunnelId int    `json:"tunnel-id"`
	Exists   bool   `json:"exists"`
}

func FindExistingRemoteInterface(ethName string, containerPid int) (string, int, bool, error) {
	stdout, err := execSelf("find-existing-remote-interface",
		"--eth-name="+ethName,
		"--container-pid="+strconv.Itoa(containerPid))
	if err != nil {
		return "", 0, false, err
	}

	existing := existing{}
	err = json.Unmarshal(stdout, &existing)
	fmt.Println("Discovered link to", existing.FabricIp, "via", existing.TunnelId)
	return existing.FabricIp, existing.TunnelId, existing.Exists, err
}

func findExistingRemoteInterface(ethName string, containerPid int) (*existing, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//get container handle
	containerNs, err := netns.GetFromPid(containerPid)
	if err != nil {
		return nil, err
	}
	defer containerNs.Close()
	containerHandle, err := netlink.NewHandleAt(containerNs)
	if err != nil {
		return nil, err
	}

	if link, err := containerHandle.LinkByName(ethName); err == nil {
		if link, isVxlan := link.(*netlink.Vxlan); isVxlan {
			return &existing{FabricIp: link.Group.String(), TunnelId: link.VxlanId, Exists: true}, nil
		}
	}
	return nil, nil
}
