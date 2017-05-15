package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/khagerma/cord-networking/network"
)

func main() {

	//if err := resolver.GetPhysicalInterface(); err != nil {
	//	panic(err)
	//}

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	//1. start listening for changes
	events, errs := cli.Events(context.Background(), types.EventsOptions{})

	//2. initialize the network
	net := network.New()

	//3. apply changes that happened while network was initializing,
	//   then continue listening for events
	for {
		select {
		case event := <-events:

			fmt.Println(event.Type, event.Action, string(event.Actor.ID[0:12]), event.From, event.Status, event.Actor.Attributes)

			if event.Type == "container" && event.Action == "start" {
				if err := net.UpdateContainer(event.Actor.ID); err != nil {
					fmt.Println(err)
				}
			}

			if event.Type == "container" && event.Action == "die" {
				net.RemoveContainer(event.Actor.ID)
			}

			break
		case err := <-errs:
			fmt.Println(err)
			break
		}
	}
}

// docker run --name=test -d --label=com.opencord.network.graph={\"links\":{\"eth0\":\"UUID-1\"}} alpine sleep 100000
