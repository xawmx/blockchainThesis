package anon

import (
	"go.dedis.ch/kyber"
)

// Suite represents the set of functionalities needed by the package anon.
type Suite interface {
	kyber.Group
	kyber.Encoding
	kyber.XOFFactory
	kyber.Random
}
