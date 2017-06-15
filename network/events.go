package network

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"context"
	"fmt"
)

func (net *Network) FireEvent(linkId graph.LinkID) {
	net.eventBus <- linkId
}

func (net *Network) listenEvents() {
	for true {
		select {
		case linkId := <-net.eventBus:
			fmt.Println("Event for link:", linkId)

			linkMap := net.graph.GetRelatedTo(linkId)

			//setup if the link exists
			if err := net.tryCreateContainerLink(linkMap, linkId); err != nil {
				fmt.Println(err)
			}

			//teardown if the link does not exist
			if err := net.tryCleanupContainerLink(linkMap, linkId); err != nil {
				fmt.Println(err)
			}

			//try to setup connection to container
			if err := net.tryCreateRemoteLink(linkMap, linkId); err != nil {
				fmt.Println(err)
			}

			if err := net.tryCleanupRemoteLink(linkMap, linkId); err != nil {
				fmt.Println(err)
			}

		}
	}
}

//tryCreateContainerLink checks if the linkMap contains two containers, and if so, ensures interfaces are set up
func (net *Network) tryCreateContainerLink(nets []graph.ContainerNetwork, linkId graph.LinkID) error {
	if len(nets) == 2 {
		fmt.Printf("Should link (linkId: %s):\n  %s in %s\n  %s in %s\n",
			linkId,
			nets[0].GetIfaceFor(linkId), nets[0].ContainerId[0:12],
			nets[1].GetIfaceFor(linkId), nets[1].ContainerId[0:12])

		pid0, err := net.getContainerPid(nets[0].ContainerId)
		if err != nil {
			return err
		}
		pid1, err := net.getContainerPid(nets[1].ContainerId)
		if err != nil {
			return err
		}

		if err := resolver.SetupLocalContainerLink(nets[0].GetIfaceFor(linkId), pid0, nets[1].GetIfaceFor(linkId), pid1); err != nil {
			return err
		}
	}
	return nil
}

//tryCleanupContainerLink checks if the linkMap contains only one container, and if so, ensures interfaces are deleted
func (net *Network) tryCleanupContainerLink(nets []graph.ContainerNetwork, linkId graph.LinkID) error {
	if len(nets) == 1 {
		fmt.Printf("Should clean (linkId: %s)\n", linkId)
		containerPid, err := net.getContainerPid(nets[0].ContainerId)
		if err != nil {
			return err
		}
		if removed, err := resolver.DeleteContainerInterface(nets[0].GetIfaceFor(linkId), containerPid); err != nil {
			return err
		} else if removed {
			fmt.Println("Removing:\n  ", nets[0].GetIfaceFor(linkId), "in", nets[0].ContainerId[0:12])
		}
	}
	return nil
}

//tryCreateRemoteLink checks if the linkMap contains one container, and if so, tries to set up a remote link
func (net *Network) tryCreateRemoteLink(nets []graph.ContainerNetwork, linkId graph.LinkID) error {
	if len(nets) == 1 {
		fmt.Printf("Should link remote (linkId: %s):\n  %s in %s\n", linkId, nets[0].GetIfaceFor(linkId), nets[0].ContainerId[0:12])
		if setup, err := net.remote.TryConnect(linkId); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Remote setup?:", setup)
		}
	}
	return nil
}

//tryCreateRemoteLink checks if the linkMap contains one container, and if so, tries to set up a remote link
func (net *Network) tryCleanupRemoteLink(nets []graph.ContainerNetwork, linkId graph.LinkID) error {
	if len(nets) != 1 {
		fmt.Printf("Should clean remotes (linkId: %s)\n", linkId)
		deleted := net.remote.TryCleanup(linkId)
		fmt.Println("Remote cleaned up?:", deleted)
	}
	return nil
}

func (net *Network) getContainerPid(containerId graph.ContainerID) (int, error) {
	if container, err := net.client.ContainerInspect(context.Background(), string(containerId)); err != nil {
		return 0, err
	} else {
		return container.State.Pid, nil
	}
}
