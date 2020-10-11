package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"machinemax/internal/generator"
	"machinemax/internal/processor"
	"machinemax/internal/registrar"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Error error `json:"error"`
}

type Server struct {
	registrar registrar.Registrar
	batchSize int
}

type Options struct {
	Registrar registrar.Registrar
	BatchSize int
}

func NewServer(opts ...func(*Options)) http.Handler {
	defaultOptions := &Options{
		Registrar: registrar.NewLoRaWANRegistrar(func(options *registrar.LoRaWANOptions) {
			options.AcceptAlreadyRegistered = true
		}),
		BatchSize: 100,
	}

	for _, fn := range opts {
		fn(defaultOptions)
	}

	s := &Server{
		registrar: defaultOptions.Registrar,
		batchSize: defaultOptions.BatchSize,
	}

	router := chi.NewRouter()
	router.Get("/ids/{idempotencyKey:[a-fA-F0-9]{8,16}}", s.IDHandler)

	return router
}

func (s *Server) IDHandler(w http.ResponseWriter, r *http.Request) {

	// Get idempotency key as hex and convert it to a usable seed value.
	// There should never be an error as the handler param requires
	// a hex value between 8 and 16, but we check in case the route
	// has been modified.
	ik := chi.URLParam(r, "idempotencyKey")
	seed, err := strconv.ParseInt(ik, 16, 64)
	if err != nil {
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Message: "could not parse idempotency key",
			Error: err,
		})
		return
	}

	rr := rand.New(rand.NewSource(seed))
	g := generator.NewGenerator(s.batchSize, generator.RandomReader(rr.Read))

	batch, err := g.Generate()
	if err != nil {
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Message: "could not generate ID batch",
			Error: err,
		})
		return
	}

	p := processor.NewProcessor(s.registrar, 10)

	// We don't need our error channel here as if there are any errors
	// registering ID's, a subsequent request may succeed in registering.
	// Because we are using an idempotent seed, this won't be an issue.
	idChan, _ := p.Process(r.Context(), batch)

	var registeredBatch generator.Batch
	for id := range idChan {
		registeredBatch = append(registeredBatch, id)
	}

	// Order the batch so our output is consistent across requests.
	// This could be unordered due to the unpredictable scheduling
	// of goroutines.
	sort.Slice(registeredBatch, func(i, j int) bool {
		return registeredBatch[i].String() < registeredBatch[j].String()
	})

	_ = json.NewEncoder(w).Encode(registeredBatch)
}
