package utils

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func Connect(host string, port int) bool {
	if c, err := net.Dial("tcp", host+`:`+strconv.Itoa(port)); err == nil {
		defer c.Close()
		return true
	} else {
		return false
	}
}

func WaitForPortAttached(port int) {
	fmt.Printf(`Waiting for port %d`, port)

	for !Connect(`localhost`, port) {
		fmt.Print(`.`)
		time.Sleep(time.Second)
	}
}
