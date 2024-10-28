// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package subtle

import (
	"crypto/cipher"
	"fmt"

	internalaead "github.com/tink-crypto/tink-go/v2/internal/aead"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
	"github.com/tink-crypto/tink-go/v2/tink"
)

const (
	// AESGCMIVSize is the acceptable IV size defined by RFC 5116.
	AESGCMIVSize = 12
	// AESGCMTagSize is the acceptable tag size defined by RFC 5116.
	AESGCMTagSize = 16

	maxIntPlaintextSize = maxInt - AESGCMIVSize - AESGCMTagSize
)

// AESGCM is an implementation of AEAD interface.
type AESGCM struct {
	cipher cipher.AEAD
}

// Assert that AESGCM implements the AEAD interface.
var _ tink.AEAD = (*AESGCM)(nil)

// NewAESGCM returns an AESGCM instance, where key is the AES key with length
// 16 bytes (AES-128) or 32 bytes (AES-256).
func NewAESGCM(key []byte) (*AESGCM, error) {
	c, err := internalaead.NewAESGCMCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESGCM{cipher: c}, nil
}

// Encrypt encrypts plaintext with associatedData. The returned ciphertext
// contains both the IV used for encryption and the actual ciphertext.
//
// Note: The crypto library's AES-GCM implementation always returns the
// ciphertext with an AESGCMTagSize (16-byte) tag.
func (a *AESGCM) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	if err := internalaead.CheckPlaintextSize(uint64(len(plaintext))); err != nil {
		return nil, err
	}
	iv := random.GetRandomBytes(AESGCMIVSize)
	dst := make([]byte, 0, len(iv)+len(plaintext)+a.cipher.Overhead())
	dst = append(dst, iv...)
	return a.cipher.Seal(dst, iv, plaintext, associatedData), nil
}

// Decrypt decrypts ciphertext with associatedData.
func (a *AESGCM) Decrypt(ciphertext, associatedData []byte) ([]byte, error) {
	if len(ciphertext) < AESGCMIVSize+AESGCMTagSize {
		return nil, fmt.Errorf("ciphertext with size %d is too short", len(ciphertext))
	}
	iv := ciphertext[:AESGCMIVSize]
	return a.cipher.Open(nil, iv, ciphertext[AESGCMIVSize:], associatedData)
}
