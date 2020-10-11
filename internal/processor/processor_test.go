package processor

import (
	"context"
	"errors"
	"fmt"
	"machinemax/internal/generator"
	"machinemax/internal/registrar"
	"net/http"
	"runtime"
	"testing"
	"time"
)

func TestProcessor_Process(t *testing.T) {
	p := NewProcessor(registrar.NewLoRaWANRegistrar(registrar.MockLoRaWANClient), 1)

	want := "DEADBEEFDEADBEEF"
	h1, _ := generator.NewIDFromHex(want)
	batch := []generator.ID{h1}

	idChan, _ := p.Process(context.Background(), batch)

	select {
	case id := <- idChan:
		if id.String() != want {
			t.Errorf("Processor.Process = %v, want %v", id.String(), want)
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("test timed out after 10 milliseconds")
	}
}

// Quick and dirty response modifier
type rtf func(r *http.Request) *http.Response

func (f rtf) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

func TestProcessor_Process_ErrorRegistering(t *testing.T) {
	p := NewProcessor(registrar.NewLoRaWANRegistrar(func(options *registrar.LoRaWANOptions) {
		options.Client = &http.Client{
			Transport: rtf(func(r *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 422,
				}
			}),
		}
	}), 1)

	want := "DEADBEEFDEADBEEF"
	h1, _ := generator.NewIDFromHex(want)
	batch := []generator.ID{h1}

	_, errChan := p.Process(context.Background(), batch)

	select {
	case err := <- errChan:
		msg := fmt.Sprintf("devEUI %s has already been used", want)
		if err.Error() != msg {
			t.Errorf("Processor.Process = %v, want %v", err.Error(), msg)
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("test timed out after 10 milliseconds")
	}
}

func Test_mergeErrChan(t *testing.T) {
	// Limit to 1 concurrent process to ensure test correctness, then
	// defer to usual value for remaining tests. This is only here
	// because of the simple way of testing channel values in this test.
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(0))
	runtime.GOMAXPROCS(1)

	e1 := make(chan error, 1)
	e2 := make(chan error, 1)

	e1 <- errors.New("hello")
	e2 <- errors.New("world")

	e := mergeErrChan(e1, e2)

	if (<-e).Error() != "hello" || (<-e).Error() != "world" {
		t.Errorf("merged error channel did not return hello and world")
	}
}

func Test_mergeIDChan(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(0))
	runtime.GOMAXPROCS(1)

	g1 := make(chan generator.ID, 1)
	g2 := make(chan generator.ID, 1)

	h1, _ := generator.NewIDFromHex("DEADBEEFDEADBEEF")
	h2, _ := generator.NewIDFromHex("AAAAAAAAFFFFFFFF")

	g1 <- h1
	g2 <- h2

	g := mergeIDChan(g1, g2)

	if (<-g).String() != "DEADBEEFDEADBEEF" || (<-g).String() != "AAAAAAAAFFFFFFFF" {
		t.Errorf("merged ID channel did not return DEADBEEFDEADBEEF and AAAAAAAAFFFFFFFF")
	}
}