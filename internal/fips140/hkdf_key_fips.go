//go:build go1.24

package fips140

import (
	"crypto/hkdf"
	"fmt"
	"hash"
)

func ComputeHKDF[H hash.Hash](hashFunc func() H, secret, salt, info []byte, length uint32) ([]byte, error) {
	result, err := hkdf.Key(hashFunc, secret, salt, string(info), int(length))
	if err != nil {
		return nil, fmt.Errorf("compute of hkdf failed")
	}
	return result, nil
}
