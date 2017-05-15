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
					fmt.Println(err)
					break
				}
				pid1, err := net.getContainerPid(containerIds[1])
				if err != nil {
					fmt.Println(err)
					break
				}

				fmt.Println("interfaces:", ifaces)

				if err:=resolver.SetupLocalContainerLink(ifaces[0], pid0, ifaces[1], pid1); err!=nil{
					fmt.Println(err)
					break
				}
			}
		}
	}
}

func (net *network) getContainerPid(containerId ContainerID) (int, error) {
	if container, err := net.client.ContainerInspect(context.Background(), string(containerId)); err != nil {
		return 0, err
	} else {
		return container.State.Pid, nil
	}
}
