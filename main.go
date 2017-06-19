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

	fmt.Println("Try running like this:")
	fmt.Println("docker run \\")
	fmt.Println("  -d \\")
	fmt.Println("  --restart=always \\")
	fmt.Println("  --pid=host \\")
	fmt.Println("  --security-opt apparmor:unconfined \\")
	fmt.Println("  --cap-add=NET_ADMIN \\")
	fmt.Println("  --cap-add=SYS_ADMIN \\")
	fmt.Println("  --cap-add=SYS_PTRACE \\")
	fmt.Println("  -v /var/run/docker.sock:/var/run/docker.sock \\")
	fmt.Println("  spanneti")
	fmt.Println()

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
