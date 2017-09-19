package olt

import (
	"github.com/ciena/spanneti/spanneti/graph"
	"fmt"
)

type OltData struct {
	OLT         map[string]OltLink `json:"olt" spanneti:"olt"`
	containerId graph.ContainerID
}

type OltLink struct {
	STag uint16 `json:"s-tag" spanneti:"s-tag"`
	CTag uint16 `json:"c-tag"`
}

func (olt OltData) SetContainerID(containerId graph.ContainerID) graph.PluginData {
	olt.containerId = containerId
	return olt
}

func (olt OltData) GetIfaceForOLT(oltLink OltLink) string {
	for iface, id := range olt.OLT {
		if id == oltLink {
			return iface
		}
	}
	panic("linkId not found")
}

func (olt OltLink) String() string {
	return fmt.Sprintf("%d-%d", olt.STag, olt.CTag)
}
