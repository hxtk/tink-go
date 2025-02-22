//go:build !go1.24

package fips140

import (
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/hkdf"
)

func ComputeHKDF[H hash.Hash](hashFunc func() H, secret, salt, info []byte, length uint32) ([]byte, error) {
	result := make([]byte, length)
	kdf := hkdf.New(hashFunc, key, salt, info)
	n, err := io.ReadFull(kdf, result)
	if n != len(result) || err != nil {
		return nil, fmt.Errorf("compute of hkdf failed")
	}
}
