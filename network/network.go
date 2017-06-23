package network

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/remote"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"strings"
)

type Network struct {
	graph       *graph.Graph
	remote      *remote.RemoteManager
	client      *client.Client
	eventBus    chan graph.LinkID
	oltEventBus chan graph.OltLink
}

func New() *Network {
	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	net := &Network{
		graph:       graph.New(),
		client:      client,
		eventBus:    make(chan graph.LinkID),
		oltEventBus: make(chan graph.OltLink),
	}
	net.init()

	return net
}

func (net *Network) init() {
	containers, err := net.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	//this order is intentional, to avoid missing container changes during init
	//1. - build the current network graph
	netGraphs := make([]graph.ContainerNetwork, len(containers))
	for i, container := range containers {
		netGraphs[i], err = net.GetContainerNetwork(container.ID)
		if err != nil {
			fmt.Println(err)
		}
		net.graph.PushContainerChanges(netGraphs[i])
	}

	//2. - start serving requests
	net.remote, err = remote.New(net.graph, net.client, net.eventBus)
	if err != nil {
		panic(err)
	}

	//3. - start listening for graph changes
	go net.listenEvents()

	//4. - fire all the events for the now-ready network graph
	net.pushContainersEvents(netGraphs)
}

func (net *Network) UpdateContainer(containerId string) error {
	containerNet, err := net.GetContainerNetwork(containerId)
	if err != nil {
		return err
	}
	oldContainerNet := net.graph.PushContainerChanges(containerNet)
	net.pushContainerEvents(oldContainerNet, containerNet)
	return nil
}

func (net *Network) RemoveContainer(containerId string) {
	//push an empty network
	oldContainerNet := net.graph.PushContainerChanges(graph.GetEmptyContainerNetwork(containerId))
	net.pushContainerEvents(oldContainerNet)
}

func (net *Network) GetContainerNetwork(containerId string) (graph.ContainerNetwork, error) {
	container, err := net.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return graph.GetEmptyContainerNetwork(containerId), err
	}

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
		return graph.GetEmptyContainerNetwork(containerId), nil
	}

	containerNet, err := graph.ParseContainerNetork(containerId, networkData)
	if err != nil {
		//for a parse error, we assume no network.  Not a real error.
		fmt.Println(err)
	}
	return containerNet, nil
}

func (net *Network) pushContainerEvents(containerNets ...graph.ContainerNetwork) {
	net.pushContainersEvents(containerNets)
}

func (net *Network) pushContainersEvents(containerNets []graph.ContainerNetwork) {
	//build a map of all the networks
	linkIds := make(map[graph.LinkID]bool)
	OLTs := make(map[graph.OltLink]bool)
	for _, containerNet := range containerNets {
		for _, linkId := range containerNet.Links {
			linkIds[linkId] = true
		}
		for _, olt := range containerNet.OLT {
			OLTs[olt] = true
		}
	}

	for linkId := range linkIds {
		net.FireEvent(linkId)
	}
	for olt := range OLTs {
		net.FireOLTEvent(olt)
	}
}
