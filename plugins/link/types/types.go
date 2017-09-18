package types

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti/graph"
)

type LinkID string

type LinkData struct {
	Links       map[string]LinkID `json:"links" spanneti:"link"`
	ContainerID graph.ContainerID
}

func (d LinkData) SetContainerID(containerId graph.ContainerID) graph.PluginData {
	d.ContainerID = containerId
	return d
}

func (data LinkData) GetIfaceFor(linkId LinkID) string {
	for iface, id := range data.Links {
		if id == linkId {
			return iface
		}
	}
	panic("linkId not found")
}
