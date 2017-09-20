package link

import (
	"fmt"
	"sync"
	"time"
)

type resyncManager struct {
	mutex     sync.Mutex
	outOfSync map[peerID]map[linkID]bool
}

func (man *linkPlugin) resync(peerId peerID, linkId linkID) {
	if peerId == man.peerId {
		panic("unableToSync() called with self as peer.  Are you sure you want to do that?")
	}

	man.resyncManager.resyncUnsafe(peerId, linkId)
}

//unableToSync adds a link to the outOfSync list for the given peer
func (man *resyncManager) resyncUnsafe(peerId peerID, linkId linkID) {
	man.mutex.Lock()
	defer man.mutex.Unlock()

	if len(man.outOfSync) == 0 {
		go man.resyncProcess()
	}

	linkIdMap, have := man.outOfSync[peerId]
	if !have {
		linkIdMap = make(map[linkID]bool)
		man.outOfSync[peerId] = linkIdMap
	}
	linkIdMap[linkId] = true
}

func (man *resyncManager) resyncProcess() {
	isOutOfSync := true
	for isOutOfSync {
		time.Sleep(time.Second)
		fmt.Println("Attempting to resync...")

		man.mutex.Lock()
		for peerId, linkIdMap := range man.outOfSync {
			linkIds := make([]linkID, len(linkIdMap))
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
		man.mutex.Unlock()
	}
	fmt.Println("All nodes back in sync.")
}
