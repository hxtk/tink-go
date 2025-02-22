//go:build !go1.24

package fips140

func FIPSEnabled() bool {
	return false
}
