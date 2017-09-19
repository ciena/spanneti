package link

import (
	"github.com/ciena/spanneti/spanneti/graph"
)

type linkID string
type peerID string

type LinkData struct {
	Links       map[string]linkID `json:"links" spanneti:"link"`
	ContainerID graph.ContainerID
}

func (d LinkData) SetContainerID(containerId graph.ContainerID) graph.PluginData {
	d.ContainerID = containerId
	return d
}

func (data LinkData) GetIfaceFor(linkId linkID) string {
	for iface, id := range data.Links {
		if id == linkId {
			return iface
		}
	}
	panic("linkId not found")
}
