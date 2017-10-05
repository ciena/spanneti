package olt

import (
	"fmt"
	"github.com/ciena/spanneti/resolver"
	"github.com/ciena/spanneti/spanneti"
	"reflect"
)

const PLUGIN_NAME = "olt.plugin.spanneti.opencord.org"

type oltPlugin spanneti.Spanneti

func LoadPlugin(spanneti spanneti.Spanneti) {
	p := oltPlugin(spanneti)
	spanneti.LoadPlugin(
		PLUGIN_NAME,
		reflect.TypeOf(OltData{}),
		p.Start,
		p.Event,
		nil,
	)
}

func (p oltPlugin) Start() {
	// cleanup unused interfaces that may have been created beforehand
	for _, sTag := range resolver.GetSharedOLTInterfaces() {
		p.Event("s-tag", sTag)
	}
}

func (p oltPlugin) Event(key string, value interface{}) error {
	nets := p.GetRelatedTo(PLUGIN_NAME, key, value).([]OltData)

	switch key {
	case "olt":
		fmt.Println("Event for OLT:", value)

		olt := value.(OltLink)

		if len(nets) == 1 {
			if err := p.tryCreateOLTLink(nets, olt); err != nil {
				return err
			}
		}

	case "s-tag":
		fmt.Println("Event for s-tag:", value)

		sTag := value.(uint16)

		if len(nets) == 0 {
			if err := p.tryCleanupSharedOLTLink(sTag); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
