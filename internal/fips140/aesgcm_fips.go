//go:build go1.24

package fips140

import "crypto/cipher"

const (
	GCMNonceSize = 0
	GCMOverhead  = 28
)

func NewGCM(c cipher.Block) (cipher.AEAD, error) {
	return cipher.NewGCMWithRandomNonce(c)
}
