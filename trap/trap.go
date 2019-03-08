package trap

import (
	"os"
	"os/signal"
)

var (
	sigs chan os.Signal = make(chan os.Signal, 1)
)

func Trap(hook func(), signals ...os.Signal) {
	signal.Notify(sigs, signals...)

	go func() {
		<-sigs
		hook()
		os.Exit(1)
	}()
}

func WaitForInterrupt() {
	select {}
}
