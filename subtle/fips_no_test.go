//go:build !go1.24

package subtle_test

func fipsEnabled() bool {
	return false
}
