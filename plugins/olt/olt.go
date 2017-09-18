package olt

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/resolver"
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti"
	"fmt"
	"reflect"
)

type oltPlugin struct {
	spanneti spanneti.Spanneti
}

func New() *oltPlugin {
	return &oltPlugin{}
}

func (p oltPlugin) Name() string {
	return "olt.plugin.spanneti.opencord.org"
}

func (p *oltPlugin) Start(spanneti spanneti.Spanneti) {
	p.spanneti = spanneti

	// cleanup unused interfaces that may have been created beforehand
	for _, sTag := range resolver.GetSharedOLTInterfaces() {
		p.Event("s-tag", sTag)
	}
}

func (p oltPlugin) DataType() reflect.Type {
	return reflect.TypeOf(OltData{})
}

func (p oltPlugin) Event(key string, value interface{}) {
	nets := p.spanneti.GetRelatedTo(key, value).([]OltData)

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
