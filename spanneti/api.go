package spanneti

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti/graph"
	"context"
)

type Spanneti struct {
	*spanneti
	plugin string
}

func (spanneti Spanneti) GetAllData() interface{} {
	return spanneti.graph.GetAllForPlugin(
		spanneti.plugin,
		spanneti.plugins[spanneti.plugin].DataType())
}

func (spanneti Spanneti) GetRelatedTo(key string, value interface{}) interface{} {
	return spanneti.graph.GetRelatedTo(
		spanneti.plugin,
		key,
		value,
		spanneti.plugins[spanneti.plugin].DataType())
}

func (net *spanneti) GetContainerPid(containerId graph.ContainerID) (int, error) {
	if container, err := net.client.ContainerInspect(context.Background(), string(containerId)); err != nil {
		return 0, err
	} else {
		return container.State.Pid, nil
	}
}
