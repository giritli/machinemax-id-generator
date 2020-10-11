package signal

import (
	"context"
	"os"
	"os/signal"
)

// catchSignal will cancel given context if a termination signal is passed to the program
func Notify(cancelFunc context.CancelFunc) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s
	cancelFunc()
}
