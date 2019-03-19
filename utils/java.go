package utils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var EnableDebug bool = false

type JavaProps map[string]string

func (j *JavaProps) Args() []string {
	args := make([]string, 0)
	for k, v := range *j {
		args = append(args, `-D`+k+`=`+v)
	}
	return args
}

// Builds a `java` command invocation with a specific JAVA_HOME
type Java struct {
	Home       string
	executable string
}

func (j *Java) Build(properties JavaProps, args ...string) *exec.Cmd {
	if len(properties) > 0 {
		args = append(properties.Args(), args...)
	}

	cmd := exec.Command(j.executable, args...)
	cmd.Env = os.Environ() // inherit env

	if os.Getenv(`JAVA_HOME`) != j.Home { // ensure JAVA_HOME uses the configured Home
		cmd.Env = append(cmd.Env, `JAVA_HOME=`+j.Home)
	}

	Debug(`Using JAVA_HOME: %q`, j.Home)
	Debug("%s %v", j.executable, args)

	return cmd
}

func (j *Java) Verify() error {
	cmd := j.Build(nil, `-version`)

	if EnableDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	}

	return cmd.Run()
}

func (j *Java) Executable() string {
	return j.executable
}

func NewJava(javaHome string) *Java {
	java := filepath.Join(javaHome, `bin`, `java`)

	if `windows` == runtime.GOOS {
		java += `.exe`
	}

	return &Java{Home: javaHome, executable: java}
}
