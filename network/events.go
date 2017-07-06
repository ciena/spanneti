package network

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"context"
	"fmt"
)

func (net *Network) FireEvent(linkId graph.LinkID) {
	net.eventBus <- linkId
}

func (net *Network) FireOLTEvent(olt graph.OltLink) {
	net.oltEventBus <- olt
}

func (net *Network) listenEvents() {
	for true {
		select {
		case linkId := <-net.eventBus:
			fmt.Println("Event for link:", linkId)

			nets := net.graph.GetRelatedTo(linkId)

			//setup if the link exists
			if err := net.tryCreateContainerLink(nets, linkId); err != nil {
				fmt.Println(err)
			}

			//teardown if the link does not exist
			if err := net.tryCleanupContainerLink(nets, linkId); err != nil {
				fmt.Println(err)
			}

			//try to setup connection to container
			if err := net.tryCreateRemoteLink(nets, linkId); err != nil {
				fmt.Println(err)
			}

			if err := net.tryCleanupRemoteLink(nets, linkId); err != nil {
				fmt.Println(err)
			}

		case olt := <-net.oltEventBus:
			fmt.Println("Event for OLT:", olt)

			nets := net.graph.GetRelatedToOlt(olt)

			if err := net.tryCreateOLTLink(nets, olt); err != nil {
				fmt.Println(err)
			}

			if err := net.tryCleanupOLTLink(nets, olt); err != nil {
				fmt.Println(err)
			}

			sTagNets := net.graph.GetRelatedToSTag(olt.STag)

			if err := net.tryCleanupSharedOLTLink(sTagNets, olt.STag); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (net *Network) getContainerPid(containerId graph.ContainerID) (int, error) {
	if container, err := net.client.ContainerInspect(context.Background(), string(containerId)); err != nil {
		return 0, err
	} else {
		return container.State.Pid, nil
	}
}
