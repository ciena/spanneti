package olt

import (
	"github.com/ciena/spanneti/resolver"
	"fmt"
)

//tryCreateOLTLink checks if the linkMap contains two containers, and if so, ensures interfaces are set up
func (p *oltPlugin) tryCreateOLTLink(nets []OltData, olt OltLink) error {
	if len(nets) == 1 {
		fmt.Printf("Should link OLT (%s): %s in %s\n",
			olt,
			nets[0].GetIfaceForOLT(olt), nets[0].containerId[0:12])

		containerPid, err := p.GetContainerPid(nets[0].containerId)
		if err != nil {
			return err
		}

		if err := resolver.SetupOLTContainerLink(nets[0].GetIfaceForOLT(olt), containerPid, olt.STag, olt.CTag); err != nil {
			return err
		}
	}
	return nil
}

//tryCleanupOLTLink checks if the linkMap contains only one container, and if so, ensures interfaces are deleted
func (net *oltPlugin) tryCleanupOLTLink(nets []OltData, olt OltLink) error {

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
func (net *oltPlugin) tryCleanupSharedOLTLink(sTagNets []OltData, sTag uint16) error {
	if len(sTagNets) == 0 {
		fmt.Printf("Should clean shared interface (fabric.%d)\n", sTag)
		if err := resolver.DeleteSharedOLTInterface(sTag); err != nil {
			return err
		}
	}
	return nil
}
