//go:build go1.24

package subtle_test

import "crypto/fips140"

func fipsEnabled() bool {
	return fips140.Enabled()
}
