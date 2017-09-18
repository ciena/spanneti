package main

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/link"
	"bitbucket.ciena.com/BP_ONOS/spanneti/plugins/olt"
	"bitbucket.ciena.com/BP_ONOS/spanneti/resolver"
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti"
)

func main() {
	if resolver.SelfCall() {
		return
	}

	spanneti.Start(olt.New(), link.New())
}
