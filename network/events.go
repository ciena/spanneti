package network

import (
	"fmt"
)

func (net *network) FireEvent(linkId LinkID) {
	net.eventBus <- linkId
}

func (net *network) listenEvents() {
	for true {
		select {
		case linkId := <-net.eventBus:
			fmt.Println(linkId)

			linkMap := net.graph.getRelatedTo(linkId)

			if len(linkMap) == 2 {
				containerIds := []ContainerID{}
				ifaces := []string{}
				for containerId, containerNet := range linkMap {
					containerIds = append(containerIds, containerId)
					ifaces = append(ifaces, containerNet.getIfaceFor(linkId))
				}

				fmt.Printf("Should link:\n  %s in %s\n  %s in %s\n", ifaces[0], containerIds[0][0:12], ifaces[1], containerIds[1][0:12])
			}

			break
		}
	}
}
