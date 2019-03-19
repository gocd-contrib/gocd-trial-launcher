package utils

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func TryConnect(host string, port int) bool {
	dest := host + `:` + strconv.Itoa(port)
	Debug(`Attempting connection to %s`, dest)
	if c, err := net.Dial("tcp", dest); err == nil {
		defer c.Close()
		Debug(`  Success.`)
		return true
	} else {
		Debug(`  Failure.`)
		return false
	}
}

func WaitUntilPortAttached(port int) {
	fmt.Printf(`Waiting for port %d`, port)

	for !TryConnect(`localhost`, port) {
		fmt.Print(`.`)
		time.Sleep(time.Second)
	}
}
