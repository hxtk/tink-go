//go:build go1.24

package testutil

import "crypto/fips140"

func FIPSEnabled() bool {
	return fips140.Enabled()
}
