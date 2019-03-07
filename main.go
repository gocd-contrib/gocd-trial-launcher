package main

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/gocd-private/gocd-trial-launcher/trap"
	"github.com/gocd-private/gocd-trial-launcher/utils"
)

const (
	HTTP_PORT  = 8153
	HTTPS_PORT = 8154
	BIND_HOST  = `localhost`
)

var (
	baseDir    string = utils.BaseDir()
	packageDir string = filepath.Join(baseDir, `packages`)
	dataDir    string = filepath.Join(baseDir, `data`)
	servPkgDir string = filepath.Join(packageDir, `go-server`)
	agntPkgDir string = filepath.Join(packageDir, `go-agent`)

	javaHome string      = filepath.Join(packageDir, `jre`)
	java     *utils.Java = utils.NewJava(javaHome)

	serverWd string = filepath.Join(dataDir, `server`)
	agentWd  string = filepath.Join(dataDir, `agent`)
)

func cleanup() {
	utils.Out("\nCleaning up...")
}

func main() {
	trap.Trap(cleanup, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	if !utils.AllDirsExist(servPkgDir, agntPkgDir, javaHome) {
		utils.Die(1, "This GoCD demo archive is missing 1 or more dependencies in the `packages` directory.\nPlease extract a clean copy from the zip archive and try again.")
	}

	os.Setenv(`JAVA_HOME`, javaHome)

	if err := java.Verify(); err != nil {
		utils.Err("Error executing java binary at [%s].\nIt might be incompatible with your OS.\n\n  Cause: %v\n", java.Executable(), err)
	}

	if utils.TryConnect(BIND_HOST, HTTP_PORT) || utils.TryConnect(BIND_HOST, HTTPS_PORT) {
		utils.Die(1, `Both ports %d and %d must be free to run this demo`, HTTP_PORT, HTTPS_PORT)
	}

	utils.Out(`server: %s`, serverWd)
	utils.Out(`agent: %s`, agentWd)
	utils.Out(`python: %t`, utils.CommandExists(`python`))
	utils.Out(`foo: %t`, utils.CommandExists(`foo`))
	utils.OpenUrlInBrowser(`https://google.com`)

	utils.Out(`Press Ctrl-C to exit`)

	trap.WaitForSignals()
	os.Exit(1)
}
