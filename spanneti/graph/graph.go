package graph

import (
	"reflect"
	"sync"
)

type Graph struct {
	containerLookup map[string]map[string]map[interface{}]map[ContainerID]*ContainerNetwork
	containerMap    map[ContainerID]*ContainerNetwork
	mutex           sync.Mutex
}

func New() *Graph {
	return &Graph{
		containerLookup: make(map[string]map[string]map[interface{}]map[ContainerID]*ContainerNetwork),
		//linkMap:      make(map[LinkID]map[ContainerID]*ContainerNetwork),
		containerMap: make(map[ContainerID]*ContainerNetwork),
		//oltMap:       make(map[uint16]map[uint16]map[ContainerID]*ContainerNetwork),
	}
}

func (graph *Graph) PushContainerChanges(containerNet ContainerNetwork) (oldContainerNet ContainerNetwork) {
	graph.mutex.Lock()
	defer graph.mutex.Unlock()

	//remove old
	oldContainerNet = graph.removeContainerUnsafe(containerNet.ContainerId)

	//only add the new container if it has a non-empty containerNet
	if len(containerNet.PluginData) > 0 {
		graph.addContainerUnsafe(containerNet)
	}

	return
}

func (graph *Graph) addContainerUnsafe(containerNet ContainerNetwork) {
	//add netGraph to the container map
	graph.containerMap[containerNet.ContainerId] = &containerNet

	for plugin, keyMap := range containerNet.KeyValueMap() {
		for key, values := range keyMap {
			for _, value := range values {
				graph.addUnsafe(plugin, key, value, &containerNet)
			}
		}
	}
}

func (graph *Graph) addUnsafe(plugin, key string, value interface{}, containerNet *ContainerNetwork) {
	keyMap, exists := graph.containerLookup[plugin]
	if !exists {
		keyMap = make(map[string]map[interface{}]map[ContainerID]*ContainerNetwork)
		graph.containerLookup[plugin] = keyMap
	}

	valueMap, exists := keyMap[key]
	if !exists {
		valueMap = make(map[interface{}]map[ContainerID]*ContainerNetwork)
		keyMap[key] = valueMap
	}

	containerMap, exists := valueMap[value]
	if !exists {
		containerMap = make(map[ContainerID]*ContainerNetwork)
		valueMap[value] = containerMap
	}

	containerMap[containerNet.ContainerId] = containerNet
}

func (graph *Graph) removeContainerUnsafe(containerId ContainerID) ContainerNetwork {
	if oldContainerNet, have := graph.containerMap[containerId]; have {
		//remove old netGraph from container map
		delete(graph.containerMap, containerId)

		for plugin, pluginMap := range oldContainerNet.KeyValueMap() {
			for key, keyMap := range pluginMap {
				for _, value := range keyMap {
					graph.removeUnsafe(plugin, key, value, containerId)
				}
			}
		}

		return *oldContainerNet
	} else {
		return getEmptyContainerNetwork(containerId)
	}
}

func (graph *Graph) removeUnsafe(plugin, key string, value interface{}, containerId ContainerID) {
	if keyMap, have := graph.containerLookup[plugin]; have {
		if valueMap, have := keyMap[key]; have {
			if containerMap, have := valueMap[value]; have {
				delete(containerMap, containerId)
				if len(containerMap) == 0 {
					delete(valueMap, value)
				}
			}
			if len(valueMap) == 0 {
				delete(keyMap, key)
			}
		}
		if len(keyMap) == 0 {
			delete(graph.containerLookup, plugin)
		}
	}
}

func (graph *Graph) GetRelatedTo(plugin, key string, value interface{}, tipe reflect.Type) interface{} {
	graph.mutex.Lock()
	defer graph.mutex.Unlock()

	//map[plugin][key][value]containerNet.pluginData[plugin]
	relatedPluginValues := reflect.MakeSlice(reflect.SliceOf(tipe), 0, 0)
	if keyMap, have := graph.containerLookup[plugin]; have {
		if valueMap, have := keyMap[key]; have {
			if containerMap, have := valueMap[value]; have {
				for _, containerNet := range containerMap {
					relatedPluginValues = reflect.Append(relatedPluginValues, reflect.ValueOf(containerNet.PluginData[plugin]))
				}
			}
		}
	}
	return relatedPluginValues.Interface()
}

func (graph *Graph) GetAllForPlugin(plugin string, tipe reflect.Type) interface{} {
	graph.mutex.Lock()
	defer graph.mutex.Unlock()

	ret := reflect.MakeSlice(reflect.SliceOf(tipe), 0, 0)
	for _, containerNet := range graph.containerMap {
		if pluginData, have := containerNet.PluginData[plugin]; have {
			ret = reflect.Append(ret, reflect.ValueOf(pluginData))
		}
	}
	return ret.Interface()
}
