package network

import (
	"sync"
)

type graph struct {
	linkMap      map[LinkID]map[ContainerID]*ContainerNetwork
	containerMap map[ContainerID]*ContainerNetwork
	mutex        sync.Mutex
}

func newGraph() graph {
	return graph{
		linkMap:      make(map[LinkID]map[ContainerID]*ContainerNetwork),
		containerMap: make(map[ContainerID]*ContainerNetwork),
	}
}

func (graph *graph) pushContainerChanges(containerNet ContainerNetwork) (oldContainerNet ContainerNetwork) {
	graph.mutex.Lock()
	defer graph.mutex.Unlock()

	//remove old
	oldContainerNet = graph.removeContainerUnsafe(containerNet.containerId)

	//only add the new container if it has a non-empty containerNet
	if len(containerNet.Links) > 0 {
		graph.addContainerUnsafe(containerNet)
	}

	return
}

func (graph *graph) removeContainerUnsafe(containerId ContainerID) ContainerNetwork {
	if oldContainerNet, have := graph.containerMap[containerId]; have {
		//remove old netGraph from container map
		delete(graph.containerMap, containerId)

		//remove every reference to the old netGraph from this container map
		for _, linkId := range oldContainerNet.Links {
			delete(graph.linkMap[linkId], containerId)
			//if no containers reference this link, remove the link-specific container map
			if len(graph.linkMap[linkId]) == 0 {
				delete(graph.linkMap, linkId)
			}
		}
		return *oldContainerNet
	} else {
		return ContainerNetwork{containerId: containerId}
	}
}

func (graph *graph) addContainerUnsafe(containerNet ContainerNetwork) {
	//add netGraph to the container map
	graph.containerMap[containerNet.containerId] = &containerNet

	for _, linkId := range containerNet.Links {
		//create a link-specific container map if one doesn't exist
		if _, have := graph.linkMap[linkId]; !have {
			graph.linkMap[linkId] = make(map[ContainerID]*ContainerNetwork)
		}

		//add netGraph to the container map
		graph.linkMap[linkId][containerNet.containerId] = &containerNet
	}
}

func (graph *graph) getRelatedTo(linkId LinkID) map[ContainerID]ContainerNetwork {
	graph.mutex.Lock()
	defer graph.mutex.Unlock()

	relatedContainerNets := make(map[ContainerID]ContainerNetwork)
	for containerId, containerNet := range graph.linkMap[linkId] {
		relatedContainerNets[containerId] = *containerNet
	}
	return relatedContainerNets
}
