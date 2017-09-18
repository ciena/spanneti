package graph

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func ParseContainerNetork(containerId string, networkData string, pluginDataMap map[string]reflect.Type) (ContainerNetwork, error) {
	containerNet := GetEmptyContainerNetwork(containerId)
	for name, pluginData := range pluginDataMap {

		unmarshalledPluginData := reflect.New(pluginData).Interface()
		if err := json.Unmarshal([]byte(networkData), unmarshalledPluginData); err != nil {
			fmt.Println(err)
			continue
		}

		//only keep if this contains real data
		if !reflect.DeepEqual(unmarshalledPluginData, pluginData) {
			containerNet.PluginData[name] = reflect.ValueOf(unmarshalledPluginData).Elem().Interface().(PluginData).SetContainerID(ContainerID(containerId))
		}
	}
	return containerNet, nil
}

func GetEmptyContainerNetwork(containerId string) ContainerNetwork {
	return getEmptyContainerNetwork(ContainerID(containerId))
}

func getEmptyContainerNetwork(containerId ContainerID) ContainerNetwork {
	return ContainerNetwork{
		ContainerId: containerId,
		PluginData:  make(map[string]PluginData),
	}
}
