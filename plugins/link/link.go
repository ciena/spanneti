package link

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/resolver"
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti"
	"fmt"
	"reflect"
	"sync"
)

const PLUGIN_NAME = "link.plugin.spanneti.opencord.org"

type LinkManager struct {
	spanneti.Spanneti
	tunnelMan   tunnelManager
	peerId      peerID
	fabricIp    string
	resyncMutex sync.Mutex
	outOfSync   map[peerID]map[linkID]bool
}

func newLinkManager(spanneti spanneti.Spanneti) (*LinkManager, error) {
	fmt.Print("Determining fabric IP... ")
	fabricIp, err := resolver.DetermineFabricIp()
	if err != nil {
		return nil, err
	}
	fmt.Println(fabricIp)

	man := &LinkManager{
		tunnelMan: newTunnelManager(),
		peerId:    determineOwnId(),
		fabricIp:  fabricIp,
		Spanneti:  spanneti,
		outOfSync: make(map[peerID]map[linkID]bool),
	}

	return man, nil
}

func (man *LinkManager) start() {
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

	go man.runServer()
}

func LoadPlugin(spanneti spanneti.Spanneti) {
	plugin, err := newLinkManager(spanneti)
	if err != nil {
		panic(err)
	}
	spanneti.LoadPlugin(
		PLUGIN_NAME,
		plugin.start,
		plugin.event,
		reflect.TypeOf(LinkData{}),
	)
}
