package remote

import (
	"fmt"
	"github.com/khagerma/cord-networking/network/graph"
	"time"
)

//unableToSync adds a link to the outOfSync list for the given peer
func (man *RemoteManager) unableToSync(peerId peerID, linkId graph.LinkID) {
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
		linkIdMap = make(map[graph.LinkID]bool)
		man.outOfSync[peerId] = linkIdMap
	}
	linkIdMap[linkId] = true
}

func (man *RemoteManager) resyncProcess() {
	isOutOfSync := true
	for isOutOfSync {
		//TODO: use events instead of sync
		time.Sleep(time.Second)
		fmt.Println("Running resync...")

		man.resyncMutex.Lock()
		for peerId, linkIdMap := range man.outOfSync {
			linkIds := make([]graph.LinkID, len(linkIdMap))
			i := 0
			for linkId := range linkIdMap {
				linkIds[i] = linkId
				i++
			}

			fmt.Println(peerId)

			if err := man.tryResyncUnsafe(peerId, linkIds); err != nil {
				fmt.Println(err)
			} else {
				delete(man.outOfSync, peerId)
			}
		}

		if len(man.outOfSync) == 0 {
			isOutOfSync = false
		}
		man.resyncMutex.Unlock()
	}
	fmt.Println("All nodes back in sync.")
}
