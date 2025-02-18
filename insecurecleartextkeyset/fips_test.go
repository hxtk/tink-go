//go:build go1.24

package insecurecleartextkeyset_test

import "crypto/fips140"

func fipsEnabled() bool {
	return fips140.Enabled()
}
