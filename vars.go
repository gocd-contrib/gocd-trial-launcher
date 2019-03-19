package main

import (
	"os/exec"
	"path/filepath"

	"github.com/gocd-contrib/gocd-trial-launcher/utils"
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
