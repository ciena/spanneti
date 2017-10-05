package link

import (
	"fmt"
	"github.com/ciena/spanneti/resolver"
	"github.com/ciena/spanneti/spanneti"
	"reflect"
)

const PLUGIN_NAME = "link.plugin.spanneti.opencord.org"

type linkPlugin struct {
	spanneti.Spanneti
	resyncManager resyncManager
	tunnelMan     tunnelManager
	peerId        peerID
	fabricIp      string
}

func LoadPlugin(spanneti spanneti.Spanneti) {
	fabricIp, err := resolver.GetFabricIp()
	if err != nil {
		panic(err)
	}

	plugin := &linkPlugin{
		Spanneti: spanneti,
		resyncManager: resyncManager{
			outOfSync: make(map[peerID]map[linkID]bool),
		},
		tunnelMan: newTunnelManager(),
		peerId:    determineOwnId(),
		fabricIp:  fabricIp,
	}

	spanneti.LoadPlugin(
		PLUGIN_NAME,
		reflect.TypeOf(LinkData{}),
		plugin.start,
		plugin.event,
		plugin.newHttpHandler(),
	)
}

func (man *linkPlugin) start() {
	//scan for existing remote links
	for _, linkData := range man.GetAllDataFor(PLUGIN_NAME).([]LinkData) {
		containerPid, err := man.GetContainerPid(linkData.ContainerID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for ethName, linkId := range linkData.Links {
			man.tunnelMan.findExisting(linkId, ethName, containerPid)
		}
	}
}
