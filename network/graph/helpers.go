package graph

import (
	"encoding/json"
)

func ParseContainerNetork(containerId string, networkData string) (ContainerNetwork, error) {
	containerNet := GetEmptyContainerNetwork(containerId)
	err := json.Unmarshal([]byte(networkData), &containerNet)
	if err != nil {
		return GetEmptyContainerNetwork(containerId), err
	}
	return containerNet, nil
}

func GetEmptyContainerNetwork(containerId string) ContainerNetwork {
	return getEmptyContainerNetwork(ContainerID(containerId))
}

func getEmptyContainerNetwork(containerId ContainerID) ContainerNetwork {
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
