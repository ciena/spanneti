package remote

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/link/remote/peer"
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/link/types"
	"fmt"
	"time"
)

//unableToSync adds a link to the outOfSync list for the given peer
func (man *RemoteManager) unableToSync(peerId peer.PeerID, linkId types.LinkID) {
	if peerId == man.peerId {
		panic("unableToSync() called with self as peer.  Are you sure you want to do that?")
	}

	man.resyncMutex.Lock()
	defer man.resyncMutex.Unlock()

	if len(man.outOfSync) == 0 {
		go man.resyncProcess()
	}

	linkIdMap, have := man.outOfSync[peerId]
	if !have {
		linkIdMap = make(map[types.LinkID]bool)
		man.outOfSync[peerId] = linkIdMap
	}
	linkIdMap[linkId] = true
}

func (man *RemoteManager) resyncProcess() {
	isOutOfSync := true
	for isOutOfSync {
		time.Sleep(time.Second)
		fmt.Println("Attempting to resync...")

		man.resyncMutex.Lock()
		for peerId, linkIdMap := range man.outOfSync {
			linkIds := make([]types.LinkID, len(linkIdMap))
			i := 0
			for linkId := range linkIdMap {
				linkIds[i] = linkId
				i++
			}

			if err := man.tryResyncUnsafe(peerId, linkIds); err != nil {
				fmt.Println(peerId, err)
			} else {
				delete(man.outOfSync, peerId)
				fmt.Println(peerId, "OK")
			}
		}

		if len(man.outOfSync) == 0 {
			isOutOfSync = false
		}
		man.resyncMutex.Unlock()
	}
	fmt.Println("All nodes back in sync.")
}
