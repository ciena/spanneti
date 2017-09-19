package spanneti

import (
	"github.com/ciena/spanneti/spanneti/graph"
	"context"
	"reflect"
)

func (s *spanneti) LoadPlugin(name string, startCallback func(), eventCallback func(key string, value interface{}), dataType reflect.Type) {
	if s.started {
		panic("cannot load plugins after spanneti has started")
	}

	plugin := &Plugin{
		name:          name,
		startCallback: startCallback,
		eventCallback: eventCallback,
		dataType:      dataType,
	}

	VerifyPlugin(plugin)

	s.plugins[name] = plugin
}

func (spanneti *spanneti) GetAllDataFor(plugin string) interface{} {
	if !spanneti.started {
		panic("spanneti not started")
	}

	return spanneti.graph.GetAllForPlugin(
		plugin,
		spanneti.plugins[plugin].dataType)
}

func (spanneti *spanneti) GetRelatedTo(plugin, key string, value interface{}) interface{} {
	if !spanneti.started {
		panic("spanneti not started")
	}

	return spanneti.graph.GetRelatedTo(
		plugin,
		key,
		value,
		spanneti.plugins[plugin].dataType)
}

func (spanneti *spanneti) GetContainerPid(containerId graph.ContainerID) (int, error) {
	if !spanneti.started {
		panic("spanneti not started")
	}

	if container, err := spanneti.client.ContainerInspect(context.Background(), string(containerId)); err != nil {
		return 0, err
	} else {
		return container.State.Pid, nil
	}
}
