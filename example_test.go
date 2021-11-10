package signal_test

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"gojini.dev/signal"
)

func ExampleRouter() {
	ctx := context.Background()
	router := signal.New(ctx)

	router.Handle(syscall.SIGUSR1, func(sig os.Signal) {
		fmt.Println("signal called:", sig)
	})

	go func() {
		if e := router.Start(); e != nil {
			panic(e)
		}
	}()

	// Simulate a signal
	router.Fire(syscall.SIGUSR1)

	// Sleep for a second for signal to be handled
	time.Sleep(time.Second)

	// Stop the router
	router.Stop(nil)

	// Output: signal called: user defined signal 1
}
