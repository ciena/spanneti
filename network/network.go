package network

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/khagerma/cord-networking/network/graph"
	"github.com/khagerma/cord-networking/network/remote"
)

type network struct {
	graph    *graph.Graph
	remote   *remote.RemoteManager
	client   *client.Client
	eventBus chan graph.LinkID
}

func New() *network {
	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	net := &network{
		graph:    graph.New(),
		client:   client,
		eventBus: make(chan graph.LinkID),
	}
	net.init()

	return net
}

func (net *network) init() {
	args := filters.NewArgs()
	args.Add("label", "com.opencord.network.graph")
	containers, err := net.client.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		fmt.Println(err)
	}

	//this order is intentional, to avoid missing container changes during init
	//1. - build the current network graph
	netGraphs := make([]graph.ContainerNetwork, len(containers))
	for i, container := range containers {
		netGraphs[i] = graph.ParseContainerNetwork(container.ID, container.Labels)
		net.graph.PushContainerChanges(netGraphs[i])
	}

	//2. - start serving requests
	net.remote = remote.New("temp_self_ID", net.graph)

	//3. - start listening for graph changes
	go net.listenEvents()

	//4. - fire all the events for the now-ready network graph
	net.pushContainersEvents(netGraphs)
}

func (net *network) UpdateContainer(containerId string) error {
	container, err := net.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return err
	}

	containerNet := graph.ParseContainerNetwork(container.ID, container.Config.Labels)
	oldContainerNet := net.graph.PushContainerChanges(containerNet)
	net.pushContainerEvents(oldContainerNet, containerNet)

	return nil
}

func (net *network) RemoveContainer(containerId string) {
	//push an empty network
	oldContainerNet := net.graph.PushContainerChanges(graph.ContainerNetwork{ContainerId: graph.ContainerID(containerId)})
	net.pushContainerEvents(oldContainerNet)
}

func (net *network) pushContainerEvents(containerNets ...graph.ContainerNetwork) {
	//build a map of all the networks
	linkIds := make(map[graph.LinkID]bool)
	for _, containerNet := range containerNets {
		for _, linkId := range containerNet.Links {
			linkIds[linkId] = true
		}
	}

	for linkId := range linkIds {
		net.FireEvent(linkId)
	}
}

func (net *network) pushContainersEvents(containerNets []graph.ContainerNetwork) {
	//build a map of all the networks
	linkIds := make(map[graph.LinkID]bool)
	for _, containerNet := range containerNets {
		for _, linkId := range containerNet.Links {
			linkIds[linkId] = true
		}
	}

	for linkId := range linkIds {
		net.FireEvent(linkId)
	}
}
