package spanneti

import (
	"context"
	"fmt"
	"github.com/ciena/spanneti/spanneti/graph"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	GIT_BRANCH     string
	GIT_COMMIT_NUM string
	GIT_COMMIT     string
	CHANGED        string
)

func printLogo() {
	displayBranch := ""
	if GIT_BRANCH != "master" {
		displayBranch = GIT_BRANCH + " "
	}
	fmt.Println(`                                  _   _ `)
	fmt.Println(` ___ _ __   __ _ _ __  _ __   ___| |_(_)`)
	fmt.Println("/ __| '_ \\ / _` | '_ \\| '_ \\ / _ \\ __| |")
	fmt.Println(`\__ \ |_) | (_| | | | | | | |  __/ |_| |`)
	fmt.Println(`|___/ .__/ \__,_|_| |_|_| |_|\___|\__|_|`)
	if CHANGED == "true" {
		fmt.Println(`DEV |_|`, displayBranch+"v0."+GIT_COMMIT_NUM+".x")
		fmt.Println("Base:", GIT_COMMIT)
	} else {
		fmt.Println(`    |_|`, displayBranch+"v0."+GIT_COMMIT_NUM)
		fmt.Println(GIT_COMMIT)
	}
	fmt.Println()
	//It's simple, we kill the PACketMAN
}

type Spanneti struct {
	*spanneti
}

type spanneti struct {
	graph   *graph.Graph
	client  *client.Client
	plugins map[string]*Plugin
	started bool

	outOfSyncMutex sync.Mutex
	outOfSync      map[resyncElem]bool
}

func New() Spanneti {
	printLogo()
	client, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	spanneti := &spanneti{
		graph:     graph.New(),
		client:    client,
		plugins:   make(map[string]*Plugin),
		outOfSync: make(map[resyncElem]bool),
	}

	return Spanneti{spanneti}
}

func (s *spanneti) Start() {
	if s.started {
		return
	}
	s.started = true

	//listen for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	//1. start listening for changes
	eventChan, errChan := cli.Events(context.Background(), types.EventsOptions{})

	//2. initialize the network
	s.init()

	//3. apply changes that happened while network was initializing,
	//   and continue listening for events
	eventLoop(s, eventChan, errChan, sigChan)
}

func eventLoop(spanneti *spanneti, eventChan <-chan events.Message, errChan <-chan error, sigChan <-chan os.Signal) {
	for {
		select {
		case event := <-eventChan:
			containerEvent(spanneti, event)

		case err := <-errChan:
			panic(err)

		case signal := <-sigChan:
			fmt.Println("Received signal:", signal)
			return
		}
	}
}

func containerEvent(spanneti *spanneti, event events.Message) {
	if event.Type == events.ContainerEventType {
		if len(event.Actor.ID) >= 12 {
			fmt.Println("Container event:", event.Action, string(event.Actor.ID[0:12]))
		} else {
			fmt.Println("Container event:", event.Action, string(event.Actor.ID))
		}

		switch event.Action {
		case "start":
			if err := spanneti.UpdateContainer(event.Actor.ID); err != nil {
				fmt.Println(err)
			}
		case "die":
			spanneti.RemoveContainer(event.Actor.ID)
		}
	}
}
