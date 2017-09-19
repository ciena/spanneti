package main

import (
	"github.com/ciena/spanneti/plugins/link"
	"github.com/ciena/spanneti/plugins/olt"
	"github.com/ciena/spanneti/spanneti"
)

func main() {
	spanneti := spanneti.New()

	link.LoadPlugin(spanneti)
	olt.LoadPlugin(spanneti)

	spanneti.Start()
}
