package spanneti

import (
	"context"
	"fmt"
	"github.com/ciena/spanneti/spanneti/graph"
	"github.com/docker/docker/api/types"
	"strings"
)

func (spanneti *spanneti) init() {
	containers, err := spanneti.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	//this order is intentional, to avoid missing container changes during init
	//1. - build the current network graph
	netGraphs := make([]graph.ContainerNetwork, len(containers))
	for i, container := range containers {
		var running bool
		netGraphs[i], running, err = spanneti.GetContainerNetwork(container.ID)
		if err != nil {
			fmt.Println(err)
		}
		if !running {
			netGraphs[i] = graph.GetEmptyContainerNetwork(container.ID)
		}

		spanneti.graph.PushContainerChanges(netGraphs[i])
	}

	//start plugins
	for _, plugin := range spanneti.plugins {
		plugin.startCallback()
	}

	//start serving requests
	spanneti.startServer()

	//5. - fire all the events for the now-ready network graph
	spanneti.pushContainerEvents(netGraphs...)
}

func (net *spanneti) UpdateContainer(containerId string) error {
	containerNet, _, err := net.GetContainerNetwork(containerId)
	if err != nil {
		return err
	}
	oldContainerNet := net.graph.PushContainerChanges(containerNet)
	net.pushContainerEvents(oldContainerNet, containerNet)
	return nil
}

func (net *spanneti) RemoveContainer(containerId string) graph.ContainerNetwork {
	//push an empty network
	oldContainerNet := net.graph.PushContainerChanges(graph.GetEmptyContainerNetwork(containerId))
	//fire events
	net.pushContainerEvents(oldContainerNet)
	return oldContainerNet
}

func (net *spanneti) GetContainerNetwork(containerId string) (graph.ContainerNetwork, bool, error) {
	container, err := net.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return graph.GetEmptyContainerNetwork(containerId), false, err
	}

	running := container.State.Running || container.State.Restarting || container.State.Paused
	var networkData string
	var has bool
	for _, env := range container.Config.Env {
		if strings.HasPrefix(env, "OPENCORD_NETWORK_GRAPH=") {
			networkData = strings.TrimPrefix(env, "OPENCORD_NETWORK_GRAPH=")
			has = true
			break
		}
	}
	if !has {
		networkData, has = container.Config.Labels["com.opencord.network.graph"]
	}

	if !has {
		return graph.GetEmptyContainerNetwork(containerId), running, nil
	}

	containerNet, err := graph.ParseContainerNetork(containerId, networkData, net.getPluginData())
	if err != nil {
		//for a parse error, we assume no network.  Not a real error.
		fmt.Println(err)
	}
	return containerNet, running, nil
}
