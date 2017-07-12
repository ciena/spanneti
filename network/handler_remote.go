package network

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"fmt"
)

//tryCreateRemoteLink checks if the linkMap contains one container, and if so, tries to set up a remote link
func (net *Network) tryCreateRemoteLink(nets []graph.ContainerNetwork, linkId graph.LinkID) error {
	if len(nets) == 1 {
		fmt.Printf("Should link remote (linkId: %s):\n  %s in %s\n", linkId, nets[0].GetIfaceFor(linkId), nets[0].ContainerId[0:12])
		containerPid, err := net.getContainerPid(nets[0].ContainerId)
		if err != nil {
			return err
		}
		if setup, err := net.remote.TryConnect(linkId, nets[0].GetIfaceFor(linkId), containerPid); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Setup link to remote?:", setup)
		}
	}
	return nil
}

//tryCreateRemoteLink checks if the linkMap contains one container, and if so, tries to set up a remote link
func (net *Network) tryCleanupRemoteLink(nets []graph.ContainerNetwork, linkId graph.LinkID) error {
	if len(nets) > 1 {
		fmt.Printf("Should clean remotes (linkId: %s)\n", linkId)
		containerPid, err := net.getContainerPid(nets[0].ContainerId)
		if err != nil {
			return err
		}
		deleted := net.remote.TryCleanup(linkId, nets[0].GetIfaceFor(linkId), containerPid)
		fmt.Println("Cleaned up links to remotes?:", deleted)
	}
	return nil
}

func (net *Network) cleanInterfaces(containerNet graph.ContainerNetwork) error {
	containerPid, err := net.getContainerPid(containerNet.ContainerId)
	if err != nil {
		return err
	}
	//remote OLT links
	for ethName := range containerNet.OLT {
		if _, err := resolver.DeleteContainerOltInterface(ethName, containerPid); err != nil {
			return err
		}
	}
	//remove remote links
	for ethName, linkId := range containerNet.Links {
		net.remote.TryCleanup(linkId, ethName, containerPid)
	}

	for ethName := range containerNet.Links {
		if _, err := resolver.DeleteContainerPeerInterface(ethName, containerPid); err != nil {
			return err
		}
	}
	return nil
}
