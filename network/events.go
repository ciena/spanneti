package network

import (
	"context"
	"fmt"
	"github.com/khagerma/cord-networking/network/resolver"
)

func (net *network) FireEvent(linkId LinkID) {
	net.eventBus <- linkId
}

func (net *network) listenEvents() {
	for true {
		select {
		case linkId := <-net.eventBus:
			fmt.Println("Event for link:", linkId)

			linkMap := net.graph.getRelatedTo(linkId)

			//setup if the link exists
			if err := net.tryCreateContainerLink(linkMap, linkId); err != nil {
				fmt.Println(err)
				break
			}

			//teardown if the link does not exist
			if err := net.tryCleanupContainerLink(linkMap, linkId); err != nil {
				fmt.Println(err)
				break
			}

		}
	}
}

//tryCreateContainerLink checks if the linkMap contains two containers, and if so, ensures interfaces are set up
func (net *network) tryCreateContainerLink(linkMap map[ContainerID]ContainerNetwork, linkId LinkID) error {
	if len(linkMap) == 2 {
		containerIds := []ContainerID{}
		ifaces := []string{}
		for containerId, containerNet := range linkMap {

			containerIds = append(containerIds, containerId)
			ifaces = append(ifaces, containerNet.getIfaceFor(linkId))
		}

		fmt.Printf("Should link:\n  %s in %s\n  %s in %s\n", ifaces[0], containerIds[0][0:12], ifaces[1], containerIds[1][0:12])

		pid0, err := net.getContainerPid(containerIds[0])
		if err != nil {
			return err
		}
		pid1, err := net.getContainerPid(containerIds[1])
		if err != nil {
			return err
		}

		if err := resolver.SetupLocalContainerLink(ifaces[0], pid0, ifaces[1], pid1); err != nil {
			return err
		}
	}
	return nil
}

//tryCleanupContainerLink checks if the linkMap contains only one container, and if so, ensures interfaces are deleted
func (net *network) tryCleanupContainerLink(linkMap map[ContainerID]ContainerNetwork, linkId LinkID) error {
	if len(linkMap) == 1 {
		for containerId, containerNet := range linkMap {
			iface := containerNet.getIfaceFor(linkId)
			containerPid, err := net.getContainerPid(containerId)
			if err != nil {
				return err
			}
			if removed, err := resolver.DeleteContainerInterface(iface, containerPid); err != nil {
				return err
			} else if removed {
				fmt.Println("Removing:\n  ", iface, "in", containerId[0:12])
			}
		}
	}
	return nil
}

func (net *network) getContainerPid(containerId ContainerID) (int, error) {
	if container, err := net.client.ContainerInspect(context.Background(), string(containerId)); err != nil {
		return 0, err
	} else {
		return container.State.Pid, nil
	}
}
