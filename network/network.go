package network

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/graph"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/remote"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network/resolver"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"strings"
)

type Network struct {
	graph        *graph.Graph
	remote       *remote.RemoteManager
	client       *client.Client
	eventBus     chan graph.LinkID
	oltEventBus  chan graph.OltLink
	sTagEventBus chan uint16
}

func New() *Network {
	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	net := &Network{
		graph:        graph.New(),
		client:       client,
		eventBus:     make(chan graph.LinkID),
		oltEventBus:  make(chan graph.OltLink),
		sTagEventBus: make(chan uint16),
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
		var running bool
		netGraphs[i], running, err = net.GetContainerNetwork(container.ID)
		if err != nil {
			fmt.Println(err)
		}
		if !running {
			netGraphs[i] = graph.GetEmptyContainerNetwork(container.ID)
		}

		net.graph.PushContainerChanges(netGraphs[i])
	}

	//2. - start serving requests
	net.remote, err = remote.New(net.graph, net.client, net.eventBus, netGraphs)
	if err != nil {
		panic(err)
	}

	//3. - start listening for graph changes
	go net.listenEvents()

	//4. - cleanup unused interfaces that may have been created beforehand
	for _, sTag := range resolver.GetSharedOLTInterfaces() {
		net.FireSTagEvent(sTag)
	}

	//5. - fire all the events for the now-ready network graph
	net.pushContainersEvents(netGraphs)
}

func (net *Network) UpdateContainer(containerId string) error {
	containerNet, _, err := net.GetContainerNetwork(containerId)
	if err != nil {
		return err
	}
	oldContainerNet := net.graph.PushContainerChanges(containerNet)
	net.pushContainerEvents(oldContainerNet, containerNet)
	return nil
}

func (net *Network) RemoveContainer(containerId string) graph.ContainerNetwork {
	//push an empty network
	oldContainerNet := net.graph.PushContainerChanges(graph.GetEmptyContainerNetwork(containerId))
	//fire events
	net.pushContainerEvents(oldContainerNet)
	return oldContainerNet
}

func (net *Network) GetContainerNetwork(containerId string) (graph.ContainerNetwork, bool, error) {
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

	containerNet, err := graph.ParseContainerNetork(containerId, networkData)
	if err != nil {
		//for a parse error, we assume no network.  Not a real error.
		fmt.Println(err)
	}
	return containerNet, running, nil
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
