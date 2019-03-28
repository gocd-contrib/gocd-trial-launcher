package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gocd-contrib/gocd-trial-launcher/gocd"
	"github.com/gocd-contrib/gocd-trial-launcher/trap"
	"github.com/gocd-contrib/gocd-trial-launcher/utils"
)

var (
	dbgFlg = flag.Bool(`X`, false, `Enables debug output`)
	verFlg = flag.Bool(`version`, false, `Displays versions and exits`)
)

func main() {
	flag.Parse()
	utils.EnableDebug = *dbgFlg

	if *verFlg {
		utils.Die(0, versionInfo())
	}

	trap.Trap(cleanup, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	if !utils.AllDirsExist(servPkgDir, agntPkgDir, javaHome) {
		utils.Die(1, "This GoCD test drive archive is missing 1 or more dependencies in the `packages` directory.\nPlease extract a clean copy from the downloaded archive and try again.")
	}

	utils.Debug(`Setting JAVA_HOME: %q`, javaHome)
	os.Setenv(`JAVA_HOME`, javaHome)

	if err := java.Verify(); err != nil {
		utils.Err("Error executing java binary at [%s].\nIt might be incompatible with your OS.\n\n  Cause: %v\n", java.Executable(), err)
	}

	if utils.TryConnect(gocd.BIND_HOST, gocd.HTTP_PORT) || utils.TryConnect(gocd.BIND_HOST, gocd.HTTPS_PORT) {
		utils.Die(1, `Both ports %d and %d must be free to run this test drive.`, gocd.HTTP_PORT, gocd.HTTPS_PORT)
	}

	if err := utils.MkdirP(serverWd, agentWd); err != nil {
		utils.Die(1, "Could not create a local data directory; please check your file permissions:\n  Cause: %v", err)
	}

	gocd.PrintLogo()

	var err error
	serverCmd, err = gocd.StartServer(java, serverWd, filepath.Join(servPkgDir, "go.jar"))

	if err != nil {
		utils.Err("Could not start the GoCD server.\n  Cause: %v", err)
		cleanup()
	}

	utils.WaitUntilPortAttached(gocd.HTTPS_PORT, `Waiting for GoCD to bootstrap`)

	utils.Out("\n")
	utils.Out("Server log directory: %q", filepath.Join(serverWd, `logs`))
	utils.Out("Agent log directory:  %q", filepath.Join(agentWd, `logs`))
	utils.Out("All data written to:  %q", dataDir)

	utils.OpenUrlInBrowser(gocd.WEB_URL + `?redirect_to=` + url.QueryEscape(`https:/tiny.cc/gocd-ux`))

	utils.Out("")
	utils.WaitUntilResponseSuccess(gocd.WEB_URL, `GoCD agent is awaiting communication with GoCD server`)
	agentCmd, err = gocd.StartAgent(java, agentWd, filepath.Join(agntPkgDir, "agent-bootstrapper.jar"))

	if err != nil {
		utils.Err("Could not start the GoCD agent.\n  Cause: %v", err)
		cleanup()
	}

	utils.Out("\nPress Ctrl-C to exit")

	trap.WaitForInterrupt()
}

func versionInfo() string {
	return fmt.Sprintf(`run-gocd %s %s (%s)`, Version, Platform, GitCommit)
}
