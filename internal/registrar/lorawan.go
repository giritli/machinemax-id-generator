package registrar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"machinemax/internal/generator"
	"net/http"
	"net/url"
)

type LoRaWANPayload struct {
	ID generator.ID `json:"deveui"`
}

type LoRaWANRegistrar struct {
	url url.URL
	client *http.Client

	// If true, treat already registered ID's as a new ID. Useful
	// When registering an idempotent request of ID batches.
	acceptAlreadyRegistered bool
}

type LoRaWANOptions struct {
	Client *http.Client
	URL    url.URL
	AcceptAlreadyRegistered bool
}

// Quick and dirty response modifier
type rtf func(r *http.Request) *http.Response

func (f rtf) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

// MockLoRaWANClient will create a mock http client which always
// returns 200 status code. This way we can utilise the LoRaWAN
// registrar without actually calling the endpoint.
func MockLoRaWANClient(options *LoRaWANOptions) {
	options.Client = &http.Client{
		Transport: rtf(func(r *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
			}
		}),
	}
}

// NewLoRaWANRegistrar will initiate a new registrar which by default will
// make requests to the LoRaWAN endpoint. Can be used to inject mock http
// clients and different endpoint URL's.
func NewLoRaWANRegistrar(opts ...func(options *LoRaWANOptions)) *LoRaWANRegistrar {
	options := &LoRaWANOptions{
		Client: &http.Client{},
		URL: url.URL{
			Scheme:     "https",
			Host:       "europe-west1-machinemax-dev-d524.cloudfunctions.net",
			Path:       "sensor-onboarding-sample",
		},
	}

	// Loop through any option function modifiers to overwrite default
	// registrar options.
	for _, fn := range opts {
		fn(options)
	}

	return &LoRaWANRegistrar{
		url: options.URL,
		client: options.Client,
		acceptAlreadyRegistered: options.AcceptAlreadyRegistered,
	}
}

// Register will register a given ID against the LoRaWAN service
func (l *LoRaWANRegistrar) Register(id generator.ID) error {
	payload := LoRaWANPayload{
		ID: id,
	}

	// This error case should never be hit but we will handle
	// just in case something goes horribly wrong.
	jsn, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create new request to service endpoint
	req, err := http.NewRequest(http.MethodPost, l.url.String(), bytes.NewReader(jsn))
	if err != nil {
		return err
	}

	// Execute request using given client
	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		return nil
	case 422:
		if l.acceptAlreadyRegistered {
			return nil
		}

		return errors.New(fmt.Sprintf("devEUI %s has already been used", id))
	default:
		return errors.New("unknown response code returned")
	}
}

