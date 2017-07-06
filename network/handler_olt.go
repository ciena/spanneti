package network

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"fmt"
)

//tryCreateOLTLink checks if the linkMap contains two containers, and if so, ensures interfaces are set up
func (net *Network) tryCreateOLTLink(nets []graph.ContainerNetwork, olt graph.OltLink) error {
	if len(nets) == 1 {
		fmt.Printf("Should link OLT (%s): %s in %s\n",
			olt,
			nets[0].GetIfaceForOLT(olt), nets[0].ContainerId[0:12])

		containerPid, err := net.getContainerPid(nets[0].ContainerId)
		if err != nil {
			return err
		}

		if err := resolver.SetupOltContainerLink(nets[0].GetIfaceForOLT(olt), containerPid, olt.STag, olt.CTag); err != nil {
			return err
		}
	}
	return nil
}

//tryCleanupOLTLink checks if the linkMap contains only one container, and if so, ensures interfaces are deleted
func (net *Network) tryCleanupOLTLink(nets []graph.ContainerNetwork, olt graph.OltLink) error {

	//TODO: how to do this?
	//if len(nets) != 1 {
	//	fmt.Printf("Should clean OLT (%s)\n", olt)
	//	containerPid, err := net.getContainerPid(nets[0].ContainerId)
	//	if err != nil {
	//		return err
	//	}
	//	if removed, err := resolver.DeleteContainerOltInterface(nets[0].GetIfaceForOLT(olt), containerPid); err != nil {
	//		return err
	//	} else if removed {
	//		fmt.Printf("Removed OLT interface:\n  %s in %s\n", nets[0].GetIfaceForOLT(olt), nets[0].ContainerId[0:12])
	//	}
	//}
	return nil
}

//tryCleanupSharedSTagLink checks if this s-tag has zero containers, and if so, deletes the shared interface
func (net *Network) tryCleanupSharedOLTLink(sTagNets []graph.ContainerNetwork, sTag uint16) error {
	if len(sTagNets) == 0 {
		fmt.Printf("Should clean shared interface (fabric.%d)\n", sTag)
		if err := resolver.DeleteSharedOltInterface(sTag); err != nil {
			return err
		}
	}
	return nil
}
