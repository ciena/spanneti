package network

import (
	"encoding/json"
	"fmt"
)

func parseContainerNetwork(containerId string, containerLabels map[string]string) ContainerNetwork {
	data := ContainerNetwork{containerId: ContainerID(containerId)}

	value, has := containerLabels["com.opencord.network.graph"]
	if !has {
		return data
	}

	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		fmt.Println("Error while parsing network graph:", err)
	}
	return data
}
