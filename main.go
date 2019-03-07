package main

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/gocd-private/gocd-trial-launcher/trap"
	"github.com/gocd-private/gocd-trial-launcher/utils"
)

var (
	baseDir    string = utils.BaseDir()
	packageDir string = filepath.Join(baseDir, `packages`)
	dataDir    string = filepath.Join(baseDir, `data`)

	javaHome string = filepath.Join(packageDir, `jre`)
	serverWd string = filepath.Join(dataDir, `server`)
	agentWd  string = filepath.Join(dataDir, `agent`)
)

func cleanup() {
	utils.Out("\nCleaning up...")
}

func main() {
	trap.Trap(cleanup, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	os.Setenv(`JAVA_HOME`, javaHome)
	utils.Out(`JAVA_HOME: %s`, os.Getenv(`JAVA_HOME`))
	utils.Out(`server: %s`, serverWd)
	utils.Out(`agent: %s`, agentWd)
	utils.Out(`python: %t`, utils.CommandExists(`python`))
	utils.Out(`foo: %t`, utils.CommandExists(`foo`))
	utils.OpenUrlInBrowser(`https://google.com`)

	utils.Out(`Press Ctrl-C to exit`)

	trap.WaitForSignals()
	os.Exit(1)
}
