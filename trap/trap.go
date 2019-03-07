package trap

import (
	"os"
	"os/signal"
)

var (
	sigs chan os.Signal = make(chan os.Signal, 1)
	done chan bool      = make(chan bool, 1)
)

func Trap(hook func(), signals ...os.Signal) {
	signal.Notify(sigs, signals...)

	go func() {
		<-sigs
		hook()
		done <- true
	}()
}

func WaitForSignals() {
	<-done
}
