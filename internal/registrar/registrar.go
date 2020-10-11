package registrar

import (
	"machinemax/internal/generator"
)

type Registrar interface {

	// Context was not included here as a requirement was not to drop
	// any ID's from registering if the application terminated early.
	Register(id generator.ID) error
}