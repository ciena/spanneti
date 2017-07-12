package network

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
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
	if len(nets) != 1 {
		fmt.Printf("Should clean remotes (linkId: %s)\n", linkId)
		if err := net.remote.TryCleanup(linkId); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

//func (net *Network) cleanInterfaces(containerNet graph.ContainerNetwork) error {
//	containerPid, err := net.getContainerPid(containerNet.ContainerId)
//	if err != nil {
//		return err
//	}
//	//remote OLT links
//	for ethName := range containerNet.OLT {
//		if _, err := resolver.DeleteContainerOltInterface(ethName, containerPid); err != nil {
//			return err
//		}
//	}
//	//remove remote links
//	for _, linkId := range containerNet.Links {
//		if err := net.remote.TryCleanup(linkId); err != nil {
//			return err
//		}
//	}
//
//	for ethName := range containerNet.Links {
//		if _, err := resolver.DeleteContainerPeerInterface(ethName, containerPid); err != nil {
//			return err
//		}
//	}
//	return nil
//}
