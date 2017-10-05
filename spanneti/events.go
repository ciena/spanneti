package spanneti

import (
	"github.com/ciena/spanneti/spanneti/graph"
)

func (spanneti *spanneti) pushContainerEvents(containerNets ...graph.ContainerNetwork) {
	todo := make(map[string]map[string][]interface{})
	for _, containerNet := range containerNets {
		for plugin, keyMap := range containerNet.KeyValueMap() {
			if _, exists := todo[plugin]; !exists {
				todo[plugin] = make(map[string][]interface{})
			}
			for key, values := range keyMap {
				todo[plugin][key] = append(todo[plugin][key], values...)
			}
		}
	}

	for plugin, pluginMap := range todo {
		for key, keyMap := range pluginMap {
			for _, value := range keyMap {
				spanneti.FireEvent(plugin, key, value)
			}
		}
	}
}
