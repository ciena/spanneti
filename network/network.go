package network

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

//nodeId uniquely identifies a container in the network graph

type network struct {
	graph    graph
	client   *client.Client
	eventBus chan LinkID
}

func New() *network {
	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	net := &network{
		graph:    newGraph(),
		client:   client,
		eventBus: make(chan LinkID),
	}
	//get the complete current state of the graph
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
	netGraphs := make([]ContainerNetwork, len(containers))
	for i, container := range containers {
		netGraphs[i] = parseContainerNetwork(container.ID, container.Labels)
		net.graph.pushContainerChanges(netGraphs[i])
	}

	//2. - start listening for graph changes
	go net.listenEvents()

	//3. - fire all the events for the now-ready network graph
	net.pushContainersEvents(netGraphs)
}

func (net *network) UpdateContainer(containerId string) error {
	container, err := net.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return err
	}

	containerNet := parseContainerNetwork(container.ID, container.Config.Labels)
	oldContainerNet := net.graph.pushContainerChanges(containerNet)
	net.pushContainerEvents(oldContainerNet, containerNet)

	return nil
}

func (net *network) RemoveContainer(containerId string) {
	//push an empty network
	oldContainerNet := net.graph.pushContainerChanges(ContainerNetwork{containerId: ContainerID(containerId)})
	net.pushContainerEvents(oldContainerNet)
}

func (net *network) pushContainerEvents(containerNets ...ContainerNetwork) {
	//build a map of all the networks
	linkIds := make(map[LinkID]bool)
	for _, containerNet := range containerNets {
		for _, linkId := range containerNet.Links {
			linkIds[linkId] = true
		}
	}

	for linkId := range linkIds {
		net.FireEvent(linkId)
	}
}

func (net *network) pushContainersEvents(containerNets []ContainerNetwork) {
	//build a map of all the networks
	linkIds := make(map[LinkID]bool)
	for _, containerNet := range containerNets {
		for _, linkId := range containerNet.Links {
			linkIds[linkId] = true
		}
	}

	for linkId := range linkIds {
		net.FireEvent(linkId)
	}
}
