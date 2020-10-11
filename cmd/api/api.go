package main

import (
	"context"
	"log"
	"machinemax/internal/server"
	"machinemax/internal/signal"
	"net/http"
	"sync"
	"time"
)

func main() {
	r := server.NewServer()

	ctx, cf := context.WithCancel(context.Background())
	go signal.Notify(cf)

	srv := http.Server{
		Addr:              ":8080",
		Handler:           r,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// If the server errors out with anything other than server close,
		// panic the error.
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down the server, waiting for remaining requests")

		// Create a context which times out after 1 minute. This should give
		// ample time for all remaining requests to be processed before
		// force shutting the server.
		timeoutCtx, _ := context.WithTimeout(context.Background(), 1 * time.Minute)
		if err := srv.Shutdown(timeoutCtx); err != nil {
			log.Fatal(err)
		}
	}

	wg.Wait()
}
