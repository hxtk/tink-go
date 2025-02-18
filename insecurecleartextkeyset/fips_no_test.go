//go:build !go1.24

package insecurecleartextkeyset_test

func fipsEnabled() bool {
	return false
}
