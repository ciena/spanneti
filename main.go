package main

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/network"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	fmt.Println(`                                  _   _ `)
	fmt.Println(` ___ _ __   __ _ _ __  _ __   ___| |_(_)`)
	fmt.Println("/ __| '_ \\ / _` | '_ \\| '_ \\ / _ \\ __| |")
	fmt.Println(`\__ \ |_) | (_| | | | | | | |  __/ |_| |`)
	fmt.Println(`|___/ .__/ \__,_|_| |_|_| |_|\___|\__|_|`)
	fmt.Println(`    |_| v0.01`)
	fmt.Println()
	//It's simple, we kill the PACketMAN

	//if err := resolver.GetPhysicalInterface(); err != nil {
	//	panic(err)
	//}

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
	net := network.New()

	//3. apply changes that happened while network was initializing,
	//   and continue listening for events
	eventLoop(net, eventChan, errChan, sigChan)
}

func eventLoop(net *network.Network, eventChan <-chan events.Message, errChan <-chan error, sigChan <-chan os.Signal) {
	for {
		select {
		case event := <-eventChan:
			containerEvent(net, event)

		case err := <-errChan:
			panic(err)

		case signal := <-sigChan:
			fmt.Println("Received signal:", signal)
			return
		}
	}
}

func containerEvent(net *network.Network, event events.Message) {
	if event.Type == events.ContainerEventType {
		if len(event.Actor.ID) >= 12 {
			fmt.Println("Container event:", event.Action, string(event.Actor.ID[0:12]))
		} else {
			fmt.Println("Container event:", event.Action, string(event.Actor.ID))
		}

		switch event.Action {
		case "start":
			if err := net.UpdateContainer(event.Actor.ID); err != nil {
				fmt.Println(err)
			}
		case "die":
			net.RemoveContainer(event.Actor.ID)
		}
	}
}
