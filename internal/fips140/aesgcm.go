//go:build !go1.24

package fips140

import "crypto/cipher"

const (
	GCMNonceSize = 12
	GCMOverhead  = 16
)

func NewGCM(c cipher.Block) (cipher.AEAD, error) {
	return cipher.NewGCM(c)
}
