package link

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/link/remote"
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/link/types"
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti"
	"fmt"
	"reflect"
)

type linkPlugin struct {
	spanneti.Spanneti
	remote *remote.RemoteManager
}

func New() *linkPlugin {
	return &linkPlugin{}
}

func (p linkPlugin) Name() string {
	return "link.plugin.spanneti.opencord.org"
}

func (p *linkPlugin) Start(spanneti spanneti.Spanneti) {
	p.Spanneti = spanneti

	//2. - start serving requests
	remote, err := remote.New(p.Spanneti, p)
	if err != nil {
		panic(err)
	}
	p.remote = remote
}

func (p linkPlugin) DataType() reflect.Type {
	return reflect.TypeOf(types.LinkData{})
}

func (plugin *linkPlugin) Event(key string, value interface{}) {
	switch key {
	case "link":
		fmt.Println("Event for link:", value)

		linkId := value.(types.LinkID)

		nets := plugin.GetRelatedTo(key, linkId).([]types.LinkData)

		//setup if the link exists
		if err := plugin.tryCreateContainerLink(nets, linkId); err != nil {
			fmt.Println(err)
		}

		//teardown if the link does not exist
		if err := plugin.tryCleanupContainerLink(nets, linkId); err != nil {
			fmt.Println(err)
		}

		//try to setup connection to container
		if err := plugin.tryCreateRemoteLink(nets, linkId); err != nil {
			fmt.Println(err)
		}

		if err := plugin.tryCleanupRemoteLink(nets, linkId); err != nil {
			fmt.Println(err)
		}
	}
}
