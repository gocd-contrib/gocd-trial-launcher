package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/gocd-private/gocd-trial-launcher/gocd"
	"github.com/gocd-private/gocd-trial-launcher/trap"
	"github.com/gocd-private/gocd-trial-launcher/utils"
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

	// These should be set by the linker at build time
	Version   = `devbuild`
	GitCommit = `unknown`
	Platform  = `devbuild`
)

var agentCmd *exec.Cmd
var serverCmd *exec.Cmd

func cleanup() {
	utils.Out("\nEnding GoCD test drive...")

	gocd.StopServer(serverCmd)
	gocd.StopAgent(agentCmd)

	utils.Out("Done. Removing this directory will remove all traces of the GoCD test drive from your system.")
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

	if utils.TryConnect(gocd.BIND_HOST, gocd.HTTP_PORT) || utils.TryConnect(gocd.BIND_HOST, gocd.HTTPS_PORT) {
		utils.Die(1, `Both ports %d and %d must be free to run this demo`, gocd.HTTP_PORT, gocd.HTTPS_PORT)
	}

	if err := utils.MkdirP(serverWd, agentWd); err != nil {
		utils.Die(1, "Could not create a local data directory; please check your file permissions:\n  Cause: %v", err)
	}

	gocd.PrintLogo()

	var err error
	serverCmd, err = gocd.StartServer(java, serverWd, filepath.Join(servPkgDir, "go.jar"))

	if err != nil {
		utils.Err("Could not start the GoCD server. Cause: %v", err)
		cleanup()
	}

	utils.WaitUntilPortAttached(gocd.HTTPS_PORT)

	agentCmd, err = gocd.StartAgent(java, agentWd, filepath.Join(agntPkgDir, "agent-bootstrapper.jar"))

	if err != nil {
		utils.Err("Could not start the GoCD agent. Cause: %v", err)
		cleanup()
	}

	utils.Out("\n\n")
	utils.Out("Server logs written to: %q", filepath.Join(serverWd, `logs`))
	utils.Out("Agent logs written to: %q", filepath.Join(agentWd, `logs`))
	utils.OpenUrlInBrowser(gocd.BrowserUrl())

	utils.Out(`Press Ctrl-C to exit`)

	trap.WaitForInterrupt()
}
