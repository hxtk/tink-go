// Package fips140 defines abstractions over primitives that were changed in Go 1.24
// to support the Go CMVP Cryptographic Module released as part of Go 1.24.
//
// These primitives did not exist (or existed outside of the cryptographic module) in
// Go 1.23 and therefore require separate implementations so that Tink can continue
// to work in Go 1.22 and Go 1.23 until such a time as the minimum supported Go
// version for Tink is increased to 1.24 or higher.
package fips140
