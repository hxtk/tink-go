//go:build !go1.24

// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package aesgcm

import (
	"crypto/cipher"

	"github.com/tink-crypto/tink-go/v2/internal/aead"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func newGCM(c cipher.Block) (cipher.AEAD, error) {
	return cipher.NewGCM(c)
}

// Encrypt encrypts plaintext with associatedData.
//
// The returned ciphertext is of the form:
//
//	prefix || iv || ciphertext || tag
//
// where prefix is the key's output prefix, iv is a random 12-byte IV,
// ciphertext is the encrypted plaintext, and tag is a 16-byte tag.
func (a *fullAEAD) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	if err := aead.CheckPlaintextSize(uint64(len(plaintext))); err != nil {
		return nil, err
	}
	iv := random.GetRandomBytes(ivSize)
	dst := make([]byte, 0, len(a.prefix)+len(iv)+len(plaintext)+a.cipher.Overhead())
	dst = append(dst, a.prefix...)
	dst = append(dst, iv...)
	return a.cipher.Seal(dst, iv, plaintext, associatedData), nil
}
