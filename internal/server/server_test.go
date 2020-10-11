package server

import (
	"bytes"
	"machinemax/internal/registrar"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestIDHandler(t *testing.T) {
	tests := []struct {
		name string
		batchSize int
		req *http.Request
		body []byte
	}{
		{
			"request",
			1,
			&http.Request{
				Method:           "GET",
				URL:              &url.URL{
					Path:       "/ids/DEADBEEF",
				},
			},
			[]byte(`["2D1596B0D99D7732"]` + "\n"),
		},
		{
			"request",
			2,
			&http.Request{
				Method:           "GET",
				URL:              &url.URL{
					Path:       "/ids/DEADBEEF",
				},
			},
			[]byte(`["2D1596B0D99D7732","F9F319F764C1036C"]` + "\n"),
		},
	}
	for _, tt := range tests {
		s := NewServer(func(options *Options) {
			options.Registrar = registrar.NewLoRaWANRegistrar(registrar.MockLoRaWANClient)
			options.BatchSize = tt.batchSize
		})

		t.Run(tt.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			s.ServeHTTP(resp, tt.req)
			body := resp.Body.Bytes()

			if !bytes.Equal(body, tt.body) {
				t.Errorf("Server.IDHandler = %s, want %s", body, tt.body)
			}
		})
	}
}