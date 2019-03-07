package utils

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func TryConnect(host string, port int) bool {
	if c, err := net.Dial("tcp", host+`:`+strconv.Itoa(port)); err == nil {
		defer c.Close()
		return true
	} else {
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
