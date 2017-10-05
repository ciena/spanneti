package ip

import (
	"github.com/ciena/spanneti/resolver"
	"github.com/ciena/spanneti/spanneti"
	"reflect"
)

const PLUGIN_NAME = "tenant-ip.plugin.spanneti.opencord.org"

type tenantIpPlugin struct {
	spanneti.Spanneti
	fabricIp      string
}

func LoadPlugin(spanneti spanneti.Spanneti) {
	fabricIp, err := resolver.GetFabricIp()
	if err != nil {
		panic(err)
	}

	plugin := &tenantIpPlugin{
		Spanneti: spanneti,
		fabricIp: fabricIp,
	}
	spanneti.LoadPlugin(
		PLUGIN_NAME,
		reflect.TypeOf(TenantIpData{}),
		plugin.Start,
		plugin.Event,
		plugin.newHttpHandler(),
	)
}

func (p tenantIpPlugin) Start() {
	//TODO: cleanup non-existing from ip-router???
}
