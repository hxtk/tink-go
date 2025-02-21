//go:build !go1.24

package testutil

func FIPSEnabled() bool {
	return false
}
