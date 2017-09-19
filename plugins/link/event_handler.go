package link

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/resolver"
	"fmt"
)

func (plugin *LinkManager) event(key string, value interface{}) {
	switch key {
	case "link":
		fmt.Println("Event for link:", value)

		linkId := value.(linkID)

		nets := plugin.GetRelatedTo(PLUGIN_NAME, key, linkId).([]LinkData)

		//setup if the link exists
		if err := plugin.tryCreateContainerLink(nets, linkId); err != nil {
			fmt.Println(err)
		}

		//teardown if the link does not exist
		if err := plugin.tryCleanupContainerLink(nets, linkId); err != nil {
			fmt.Println(err)
		}

		//try to setup connection to container
		if err := plugin.tryCreateRemoteLink(nets, linkId); err != nil {
			fmt.Println(err)
		}

		if err := plugin.tryCleanupRemoteLink(nets, linkId); err != nil {
			fmt.Println(err)
		}
	}
}

//tryCreateContainerLink checks if the linkMap contains two containers, and if so, ensures interfaces are set up
func (plugin *LinkManager) tryCreateContainerLink(nets []LinkData, linkId linkID) error {
	if len(nets) == 2 {
		fmt.Printf("Should link (linkId: %s):\n  %s in %s\n  %s in %s\n",
			linkId,
			nets[0].GetIfaceFor(linkId), nets[0].ContainerID[0:12],
			nets[1].GetIfaceFor(linkId), nets[1].ContainerID[0:12])

		pid0, err := plugin.GetContainerPid(nets[0].ContainerID)
		if err != nil {
			return err
		}
		pid1, err := plugin.GetContainerPid(nets[1].ContainerID)
		if err != nil {
			return err
		}

		if err := resolver.SetupLocalContainerLink(nets[0].GetIfaceFor(linkId), pid0, nets[1].GetIfaceFor(linkId), pid1); err != nil {
			return err
		}
	}
	return nil
}

//tryCleanupContainerLink checks if the linkMap contains only one container, and if so, ensures interfaces are deleted
func (plugin *LinkManager) tryCleanupContainerLink(nets []LinkData, linkId linkID) error {
	if len(nets) == 1 {
		fmt.Printf("Should clean (linkId: %s)\n", linkId)
		containerPid, err := plugin.GetContainerPid(nets[0].ContainerID)
		if err != nil {
			return err
		}
		if err := resolver.DeleteContainerPeerInterface(nets[0].GetIfaceFor(linkId), containerPid); err != nil {
			return err
		}
	}
	return nil
}

//tryCreateRemoteLink checks if the linkMap contains one container, and if so, tries to set up a remote link
func (plugin *LinkManager) tryCreateRemoteLink(nets []LinkData, linkId linkID) error {
	if len(nets) == 1 {
		fmt.Printf("Should link remote (linkId: %s):\n  %s in %s\n", linkId, nets[0].GetIfaceFor(linkId), nets[0].ContainerID[0:12])
		containerPid, err := plugin.GetContainerPid(nets[0].ContainerID)
		if err != nil {
			return err
		}
		if setup, err := plugin.tryConnect(linkId, nets[0].GetIfaceFor(linkId), containerPid); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Setup link to remote?:", setup)
		}
	}
	return nil
}

//tryCreateRemoteLink checks if the linkMap contains one container, and if so, tries to set up a remote link
func (net *LinkManager) tryCleanupRemoteLink(nets []LinkData, linkId linkID) error {
	if len(nets) != 1 {
		fmt.Printf("Should clean remotes (linkId: %s)\n", linkId)
		if err := net.tryCleanup(linkId); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

//func (net *Network) cleanInterfaces(containerNet graph.ContainerNetwork) error {
//	containerPid, err := net.getContainerPid(containerNet.ContainerId)
//	if err != nil {
//		return err
//	}
//	//remote OLT links
//	for ethName := range containerNet.OLT {
//		if _, err := resolver.DeleteContainerOltInterface(ethName, containerPid); err != nil {
//			return err
//		}
//	}
//	//remove remote links
//	for _, linkId := range containerNet.Links {
//		if err := net.remote.TryCleanup(linkId); err != nil {
//			return err
//		}
//	}
//
//	for ethName := range containerNet.Links {
//		if _, err := resolver.DeleteContainerPeerInterface(ethName, containerPid); err != nil {
//			return err
//		}
//	}
//	return nil
//}
