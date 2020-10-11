package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"machinemax/internal/generator"
	"machinemax/internal/processor"
	"machinemax/internal/registrar"
	"machinemax/internal/signal"
	"sync"
)

// How many IDs to register per batch
const IDBatchSize = 100

// How many workers to use for parallel registry
const Workers = 10

// We use log for non ID output so that we can provide context to the terminal, but
// output to StdOut for the registered ID's. Useful for piping to other processes.
func main() {
	ctx, cf := context.WithCancel(context.Background())
	go signal.Notify(cf)

	var registeredIDs []generator.ID

	main:
	// Keep registering ID's until the number of registered ID's
	// match the batch size we need. The reason we retry when an
	// ID is previously registered is because we want 100 fresh
	// ID's.
	for l := 0; l < IDBatchSize; l = len(registeredIDs) {

		// Or until the application is terminated...
		select {
		case <-ctx.Done():
			log.Println("Execution terminated early")
			break main
		default:
		}

		// Only generate remaining number of ID's. This is to account
		// for ID's that could not eb registered for whatever reason.
		remaining := IDBatchSize - l
		g := generator.NewGenerator(remaining, generator.RandomReader(rand.Read))

		log.Printf("Registering %d ID's\n", remaining)

		// If the generation of ID's fails, lets continue and try again.
		// No exit clause for this assessment. One for a future improvement.
		batch, err := g.Generate()
		if err != nil {
			log.Printf("Error generating ID's: %s\n", err.Error())
			continue
		}

		p := processor.NewProcessor(registrar.NewLoRaWANRegistrar(), Workers)
		outChan, errChan := p.Process(ctx, batch)

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			for id := range outChan {
				registeredIDs = append(registeredIDs, id)
				log.Printf("Registered: %s\n", id)
			}
		}()

		go func() {
			defer wg.Done()
			for err := range errChan {
				log.Printf("Error: %s\n", err.Error())
			}
		}()

		wg.Wait()
	}

	log.Printf("Generated %d ID's\n", len(registeredIDs))
	for _, id := range registeredIDs {
		fmt.Println(id)
	}
}