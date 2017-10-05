package spanneti

import (
	"fmt"
	"time"
)

type resyncElem struct {
	plugin string
	key    string
	value  interface{}
}

func (man *spanneti) resync(plugin, key string, value interface{}) {
	man.outOfSyncMutex.Lock()
	defer man.outOfSyncMutex.Unlock()

	if len(man.outOfSync) == 0 {
		go man.resyncProcess()
	}

	man.outOfSync[resyncElem{plugin, key, value}] = true
}

func (man *spanneti) resyncProcess() {
	isOutOfSync := true
	for isOutOfSync {
		time.Sleep(time.Second)
		fmt.Println("Attempting to resync...")

		man.outOfSyncMutex.Lock()
		for elem := range man.outOfSync {
			if err := man.plugins[elem.plugin].eventCallback(elem.key, elem.value); err != nil {
				fmt.Println(elem.plugin+":", err)
			} else {
				delete(man.outOfSync, elem)
				fmt.Println("Resync OK")
			}
		}

		if len(man.outOfSync) == 0 {
			isOutOfSync = false
		}
		man.outOfSyncMutex.Unlock()
	}
	fmt.Println("All nodes back in sync.")
}
