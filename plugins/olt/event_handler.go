package olt

import (
	"fmt"
	"github.com/ciena/spanneti/resolver"
)

//tryCreateOLTLink checks if the linkMap contains two containers, and if so, ensures interfaces are set up
func (p *oltPlugin) tryCreateOLTLink(nets []OltData, olt OltLink) error {
	fmt.Printf("Should link OLT (%s): %s in %s\n",
		olt,
		nets[0].GetIfaceForOLT(olt), nets[0].containerId[0:12])

	containerPid, err := p.GetContainerPid(nets[0].containerId)
	if err != nil {
		return err
	}

	if err := resolver.SetupContainerOLTInterface(nets[0].GetIfaceForOLT(olt), containerPid, olt.STag, olt.CTag); err != nil {
		return err
	}
	return nil
}

//tryCleanupSharedSTagLink checks if this s-tag has zero containers, and if so, deletes the shared interface
func (net *oltPlugin) tryCleanupSharedOLTLink(sTag uint16) error {
	fmt.Printf("Should clean shared interface (fabric.%d)\n", sTag)
	return resolver.DeleteSharedOLTInterface(sTag)
}
