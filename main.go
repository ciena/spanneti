package main

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/link"
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/olt"
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti"
)

func main() {
	spanneti := spanneti.New()

	link.LoadPlugin(spanneti)
	olt.LoadPlugin(spanneti)

	spanneti.Start()
}
