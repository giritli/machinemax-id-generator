package processor

import (
	"context"
	"machinemax/internal/generator"
	"machinemax/internal/registrar"
	"sync"
)

type Processor struct {
	r registrar.Registrar
	workers int
}

func NewProcessor(r registrar.Registrar, workers int) *Processor {
	return &Processor{
		r: r,
		workers: workers,
	}
}

func (p *Processor) Process(ctx context.Context, b generator.Batch) (<-chan generator.ID, <-chan error) {

	// This channel is used to convert the batch to a channel
	// of IDs.
	inChan := make(chan generator.ID)

	// This is the channel of IDs that have been successfully
	// registered by the registrar.
	outChannels := make([]<-chan generator.ID, 0)

	// This channel contains any errors that may occur
	errChannels := make([]<-chan error, 0)

	// Pipe the batch into the input channel
	go func() {
		defer close(inChan)

		for _, id := range b {
			inChan <- id
		}
	}()

	for i := 0; i < p.workers; i++ {
		outChan, errChan := p.processing(ctx, inChan)
		outChannels = append(outChannels, outChan)
		errChannels = append(errChannels, errChan)
	}

	return mergeIDChan(outChannels...), mergeErrChan(errChannels...)
}

// processing will consume the inChan and pass each result to the
// registrar to register. A ID and error channel will be returned.
func (p *Processor) processing(ctx context.Context, inChan <-chan generator.ID) (<-chan generator.ID, <-chan error) {
	outChan := make(chan generator.ID)
	errChan := make(chan error)

	go func() {
		// Close both channels once inChan is closed.
		defer close(outChan)
		defer close(errChan)

		for id := range inChan {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := p.r.Register(id)

			if err != nil {
				errChan <- err
				continue
			}

			outChan <- id
		}
	}()

	return outChan, errChan
}

// Generics would be so useful for the following two functions
func mergeErrChan(chs ...<-chan error) <-chan error {
	out := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(chs))

	consume := func(ch <-chan error) {
		for x := range ch {
			out <- x
		}
		wg.Done()
	}

	for _, ch := range chs {
		go consume(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func mergeIDChan(chs ...<-chan generator.ID) <-chan generator.ID {
	out := make(chan generator.ID)
	wg := sync.WaitGroup{}
	wg.Add(len(chs))

	consume := func(ch <-chan generator.ID) {
		for x := range ch {
			out <- x
		}
		wg.Done()
	}

	for _, ch := range chs {
		go consume(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}