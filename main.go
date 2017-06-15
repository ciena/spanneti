package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"bitbucket.ciena.com/BP_ONOS/spanneti/network"
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

			fmt.Println("Docker event:", event.Action, event.Type, string(event.Actor.ID))

			if event.Type == "container" && event.Action == "start" {
				if err := net.UpdateContainer(event.Actor.ID); err != nil {
					fmt.Println(err)
				}
			}

			if event.Type == "container" && event.Action == "die" {
				net.RemoveContainer(event.Actor.ID)
			}

		case err := <-errs:
			panic(err)
		}
	}
}
