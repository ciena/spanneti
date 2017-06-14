package graph

import (
	"encoding/json"
	"fmt"
)

func ParseContainerNetwork(containerId string, containerLabels map[string]string) ContainerNetwork {
	data := GetEmptyContainerNetwork(ContainerID(containerId))

	value, has := containerLabels["com.opencord.network.graph"]
	if !has {
		return data
	}
	fmt.Println(string(value))

	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		fmt.Println("Error while parsing network graph:", err)
	}
	return data
}

func GetEmptyContainerNetwork(containerId ContainerID) ContainerNetwork {
	return ContainerNetwork{ContainerId: ContainerID(containerId)}
}

func (contNet ContainerNetwork) GetIfaceFor(linkId LinkID) string {
	for iface, id := range contNet.Links {
		if id == linkId {
			return iface
		}
	}
	panic("linkId not found")
}
