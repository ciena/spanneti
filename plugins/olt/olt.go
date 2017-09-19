package olt

import (
	"github.com/ciena/spanneti/resolver"
	"github.com/ciena/spanneti/spanneti"
	"fmt"
	"reflect"
)

const PLUGIN_NAME = "olt.plugin.spanneti.opencord.org"

type oltPlugin spanneti.Spanneti

func LoadPlugin(spanneti spanneti.Spanneti) {
	p := oltPlugin(spanneti)
	spanneti.LoadPlugin(
		PLUGIN_NAME,
		p.Start,
		p.Event,
		reflect.TypeOf(OltData{}),
	)
}

func (p oltPlugin) Start() {
	// cleanup unused interfaces that may have been created beforehand
	for _, sTag := range resolver.GetSharedOLTInterfaces() {
		p.Event("s-tag", sTag)
	}
}

func (p oltPlugin) Event(key string, value interface{}) {
	nets := p.GetRelatedTo(PLUGIN_NAME, key, value).([]OltData)

	switch key {
	case "olt":
		fmt.Println("Event for OLT:", value)

		olt := value.(OltLink)

		if err := p.tryCreateOLTLink(nets, olt); err != nil {
			fmt.Println(err)
		}

		if err := p.tryCleanupOLTLink(nets, olt); err != nil {
			fmt.Println(err)
		}

	case "s-tag":
		fmt.Println("Event for s-tag:", value)

		sTag := value.(uint16)

		if err := p.tryCleanupSharedOLTLink(nets, sTag); err != nil {
			fmt.Println(err)
		}
	}
}
