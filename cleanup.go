package main

import (
	"github.com/gocd/gocd-trial-launcher/gocd"
	"github.com/gocd/gocd-trial-launcher/utils"
)

func cleanup() {
	utils.Out("\nEnding GoCD test drive...")

	gocd.StopServer(serverCmd)
	gocd.StopAgent(agentCmd)

	utils.Out("Done. Removing this directory will remove all traces of the GoCD test drive from your system.")
}
