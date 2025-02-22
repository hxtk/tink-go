//go:build go1.24

package fips140

import "crypto/fips140"

func FIPSEnabled() bool {
	return fips140.Enabled()
}
