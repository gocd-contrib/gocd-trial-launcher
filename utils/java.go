package utils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var Debug bool = false

type JavaProps map[string]string

func (j *JavaProps) Args() []string {
	args := make([]string, 0)
	for k, v := range *j {
		args = append(args, `-D`+k+`=`+v)
	}
	return args
}

type Java struct {
	Home string
	java string
}

func (j *Java) Run(properties JavaProps, args ...string) *exec.Cmd {
	if len(properties) > 0 {
		args = append(properties.Args(), args...)
	}

	if Debug {
		Out("%s %v", j.java, args)
	}

	cmd := exec.Command(j.java, args...)
	cmd.Env = os.Environ() // inherit env

	if os.Getenv(`JAVA_HOME`) != j.Home { // ensure JAVA_HOME uses the configured Home
		cmd.Env = append(cmd.Env, `JAVA_HOME=`+j.Home)
	}

	return cmd
}

func (j *Java) Verify() error {
	cmd := j.Run(nil, `-version`)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func (j *Java) Executable() string {
	return j.java
}

func NewJava(javaHome string) *Java {
	java := filepath.Join(javaHome, `bin`, `java`)

	if `windows` == runtime.GOOS {
		java += `.exe`
	}

	return &Java{Home: javaHome, java: java}
}
