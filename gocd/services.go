package gocd

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/gocd-contrib/gocd-trial-launcher/utils"
)

const (
	HTTP_PORT  = 8153
	HTTPS_PORT = 8154
	BIND_HOST  = `localhost`
)

var (
	WEB_URL            = `http://` + BIND_HOST + `:` + strconv.Itoa(HTTP_PORT)
	AGENT_REGISTER_URL = `https://` + BIND_HOST + `:` + strconv.Itoa(HTTPS_PORT) + `/go`
)

func StartServer(java *utils.Java, workDir, jar string) (*exec.Cmd, error) {
	configDir := filepath.Join(workDir, "config")
	configFile := filepath.Join(configDir, "cruise-config.xml")
	tmpDir := filepath.Join(workDir, "tmp")
	logDir := filepath.Join(workDir, "logs")
	logFile := filepath.Join(logDir, "stdout.log")

	if err := utils.MkdirP(configDir, tmpDir, logDir); err != nil {
		return nil, err
	}

	props := utils.JavaProps{
		"cruise.config.dir":            configDir,
		"cruise.config.file":           configFile,
		"java.io.tmpdir":               tmpDir,
		"gocd.redirect.stdout.to.file": logFile,
	}

	if err := mergeExtraProperties(props, jar); err != nil {
		return nil, err
	}

	return startJavaApp(java, "server", workDir, props, "-Xmx1024m", "-jar", jar, "-server")
}

func StartAgent(java *utils.Java, workDir, jar string) (*exec.Cmd, error) {
	tmpDir := filepath.Join(workDir, "tmp")
	logDir := filepath.Join(workDir, "logs")
	logFile := filepath.Join(logDir, "stdout.log")

	if err := utils.MkdirP(tmpDir, logDir); err != nil {
		return nil, err
	}

	props := utils.JavaProps{
		"java.io.tmpdir":               tmpDir,
		"gocd.redirect.stdout.to.file": logFile,
		"gocd.agent.log.dir":           logDir,
	}

	if err := mergeExtraProperties(props, jar); err != nil {
		return nil, err
	}

	return startJavaApp(java, "agent", workDir, props, "-Xmx256m", "-jar", jar, "-serverUrl", AGENT_REGISTER_URL)
}

func StopServer(cmd *exec.Cmd) {
	if cmd != nil {
		pidFile := filepath.Join(cmd.Dir, "server.pid")

		stopApp(cmd, pidFile, "server")
	}
}

func StopAgent(cmd *exec.Cmd) {
	if cmd != nil {
		pidFile := filepath.Join(cmd.Dir, "agent.pid")

		stopApp(cmd, pidFile, "agent")
	}
}

func mergeExtraProperties(props utils.JavaProps, jar string) error {
	utils.Debug(`Checking for extra properties for app at %q`, jar)
	path := filepath.Join(filepath.Dir(jar), `extra-props.yaml`)

	if utils.IsFile(path) {
		utils.Debug(`Reading java properties from %q`, path)

		if extras, err := utils.PropsFromYaml(path); err == nil {
			for k, v := range extras {
				utils.Debug(`Found property %s=%s`, k, v)
				props[k] = v
			}
		} else {
			utils.Debug(`Error while extracting properties from %q: %v`, path, err)
			return err
		}
	}
	return nil
}

func startJavaApp(java *utils.Java, serviceName string, workDir string, properties utils.JavaProps, args ...string) (*exec.Cmd, error) {
	cmd := java.Build(properties, args...)

	utils.EnablePgid(cmd)

	cmd.Dir = workDir
	pidFile := filepath.Join(workDir, serviceName+".pid")

	utils.Out("\nStarting the GoCD %s...", serviceName)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	utils.Debug(`%s PID: %d, writing to pidfile: %q`, serviceName, cmd.Process.Pid, pidFile)

	if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		return nil, err
	}

	return cmd, nil
}

func stopApp(cmd *exec.Cmd, pidFile, serviceName string) {
	utils.Debug(`Ending %s process %d`, serviceName, cmd.Process.Pid)

	if cmd != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
		utils.Out("Stopping GoCD %s...", serviceName)

		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			if err = cmd.Process.Kill(); err != nil {
				utils.Err("Unable to stop the GoCD test drive. See PID: %d", cmd.Process.Pid)
			}
		}
	}

	if cmd != nil {
		utils.KillPgid(cmd)
	}

	if pidFile != "" && utils.IsExist(pidFile) {
		utils.Debug(`Removing pidfile: %q`, pidFile)

		if err := os.Remove(pidFile); err != nil {
			utils.Err("Failed to remove pidfile %s.\n  Cause: %v", pidFile, err)
		}
	}
}
