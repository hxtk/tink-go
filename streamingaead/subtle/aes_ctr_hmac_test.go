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

package subtle_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/tink-crypto/tink-go/v2/internal/fips140"
	"github.com/tink-crypto/tink-go/v2/streamingaead/subtle"
)

func TestAESCTRHMACEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name               string
		keySizeInBytes     int
		tagSizeInBytes     int
		segmentSize        int
		firstSegmentOffset int
		plaintextSize      int
		chunkSize          int
	}{
		{
			name:               "small-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      20,
			chunkSize:          64,
		},
		{
			name:               "small-2",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      400,
			chunkSize:          64,
		},
		{
			name:               "small-offset-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 8,
			plaintextSize:      20,
			chunkSize:          64,
		},
		{
			name:               "small-offset-2",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        512,
			firstSegmentOffset: 8,
			plaintextSize:      400,
			chunkSize:          64,
		},
		{
			name:               "empty-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      0,
			chunkSize:          128,
		},
		{
			name:               "empty-2",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 8,
			plaintextSize:      0,
			chunkSize:          128,
		},
		{
			name:               "medium-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      1024,
			chunkSize:          128,
		},
		{
			name:               "medium-2",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      3086,
			chunkSize:          128,
		},
		{
			name:               "medium-3",
			keySizeInBytes:     32,
			tagSizeInBytes:     12,
			segmentSize:        1024,
			firstSegmentOffset: 0,
			plaintextSize:      12345,
			chunkSize:          128,
		},
		{
			name:               "large-chunks-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      1024,
			chunkSize:          4096,
		},
		{
			name:               "large-chunks-2",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      5086,
			chunkSize:          4096,
		},
		{
			name:               "large-chunks-3",
			keySizeInBytes:     32,
			tagSizeInBytes:     12,
			segmentSize:        1024,
			firstSegmentOffset: 0,
			plaintextSize:      12345,
			chunkSize:          5000,
		},
		{
			name:               "medium-offset-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 8,
			plaintextSize:      1024,
			chunkSize:          64,
		},
		{
			name:               "medium-offset-2",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        512,
			firstSegmentOffset: 20,
			plaintextSize:      3086,
			chunkSize:          256,
		},
		{
			name:               "medium-offset-3",
			keySizeInBytes:     32,
			tagSizeInBytes:     12,
			segmentSize:        1024,
			firstSegmentOffset: 10,
			plaintextSize:      12345,
			chunkSize:          5000,
		},
		{
			name:               "last-segment-full-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      216,
			chunkSize:          64,
		},
		{
			name:               "last-segment-full-2",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 16,
			plaintextSize:      200,
			chunkSize:          256,
		},
		{
			name:               "last-segment-full-3",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 16,
			plaintextSize:      440,
			chunkSize:          1024,
		},
		{
			name:               "single-byte-1",
			keySizeInBytes:     16,
			tagSizeInBytes:     12,
			segmentSize:        256,
			firstSegmentOffset: 0,
			plaintextSize:      1024,
			chunkSize:          1,
		},
		{
			name:               "single-byte-2",
			keySizeInBytes:     32,
			tagSizeInBytes:     12,
			segmentSize:        512,
			firstSegmentOffset: 0,
			plaintextSize:      5086,
			chunkSize:          1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cipher, err := subtle.NewAESCTRHMAC(ikm, "SHA256", tc.keySizeInBytes, "SHA256", tc.tagSizeInBytes, tc.segmentSize, tc.firstSegmentOffset)
			if err != nil {
				t.Errorf("cannot create cipher: %v", err)
			}
			pt, ct, err := encrypt(cipher, aad, tc.plaintextSize)
			if err != nil {
				t.Errorf("failure during encryption: %v", err)
			}
			if err := decrypt(cipher, aad, pt, ct, tc.chunkSize); err != nil {
				t.Errorf("failure during decryption: %v", err)
			}
		})
	}
}

func TestAESCTRHMACModifiedCiphertext(t *testing.T) {
	ikm, err := hex.DecodeString("000102030405060708090a0b0c0d0e0f00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	aad, err := hex.DecodeString("aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}

	const (
		keySizeInBytes     = 16
		tagSizeInBytes     = 12
		segmentSize        = 256
		firstSegmentOffset = 8
		plaintextSize      = 1024
		chunkSize          = 128
	)

	cipher, err := subtle.NewAESCTRHMAC(ikm, "SHA256", keySizeInBytes, "SHA256", tagSizeInBytes, segmentSize, firstSegmentOffset)
	if err != nil {
		t.Errorf("Cannot create a cipher: %v", err)
	}

	pt, ct, err := encrypt(cipher, aad, plaintextSize)
	if err != nil {
		t.Error(err)
	}

	t.Run("truncate ciphertext", func(t *testing.T) {
		for i := 0; i < len(ct); i += 8 {
			if err := decrypt(cipher, aad, pt, ct[:i], chunkSize); err == nil {
				t.Error("expected error")
			}
		}
	})
	t.Run("append to ciphertext", func(t *testing.T) {
		sizes := []int{1, segmentSize - len(ct)%segmentSize, segmentSize}
		for _, size := range sizes {
			ct2 := append(ct, make([]byte, size)...)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("flip bits", func(t *testing.T) {
		for i := range ct {
			ct2 := make([]byte, len(ct))
			copy(ct2, ct)
			ct2[i] ^= byte(1)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("delete segments", func(t *testing.T) {
		for i := 0; i < len(ct)/segmentSize+1; i++ {
			start, end := segmentPos(segmentSize, firstSegmentOffset, cipher.HeaderLength(), i)
			if start > len(ct) {
				break
			}
			if end > len(ct) {
				end = len(ct)
			}
			ct2 := append(ct[:start], ct[end:]...)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("duplicate segments", func(t *testing.T) {
		for i := 0; i < len(ct)/segmentSize+1; i++ {
			start, end := segmentPos(segmentSize, firstSegmentOffset, cipher.HeaderLength(), i)
			if start > len(ct) {
				break
			}
			if end > len(ct) {
				end = len(ct)
			}
			ct2 := append(ct[:end], ct[start:]...)
			if err := decrypt(cipher, aad, pt, ct2, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
	t.Run("modify aad", func(t *testing.T) {
		for i := range aad {
			aad2 := make([]byte, len(aad))
			copy(aad2, aad)
			aad2[i] ^= byte(1)
			if err := decrypt(cipher, aad2, pt, ct, chunkSize); err == nil {
				t.Errorf("expected error")
			}
		}
	})
}

func TestAESCTRHMACWithValidParameters(t *testing.T) {
	const (
		segmentSize        = 256
		firstSegmentOffset = 8
	)
	mainKey, err := hex.DecodeString(
		"000102030405060708090a0b0c0d0e0f00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		hkdfAlg        string
		tagAlg         string
		keySizeInBytes int
		tagSizeInBytes int
		noFIPS         bool
	}{
		// smallest possible key and tag sizes
		{
			hkdfAlg:        "SHA1",
			tagAlg:         "SHA1",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
			noFIPS:         true, // Uses SHA1
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA224",
			tagAlg:         "SHA224",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA384",
			tagAlg:         "SHA384",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA512",
			tagAlg:         "SHA512",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		// tagSize equal to digest size
		{
			hkdfAlg:        "SHA1",
			tagAlg:         "SHA1",
			keySizeInBytes: 16,
			tagSizeInBytes: 20,
			noFIPS:         true, // Uses SHA1
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 32,
		},
		{
			hkdfAlg:        "SHA224",
			tagAlg:         "SHA224",
			keySizeInBytes: 16,
			tagSizeInBytes: 28,
		},
		{
			hkdfAlg:        "SHA384",
			tagAlg:         "SHA384",
			keySizeInBytes: 16,
			tagSizeInBytes: 48,
		},
		{
			hkdfAlg:        "SHA512",
			tagAlg:         "SHA512",
			keySizeInBytes: 16,
			tagSizeInBytes: 64,
		},
		// hkdfAlg and tagAlg different
		{
			hkdfAlg:        "SHA1",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
			noFIPS:         true, // Uses SHA1
		},
		{
			hkdfAlg:        "SHA224",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA384",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA512",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA1",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
			noFIPS:         true, // Uses SHA1
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA224",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA384",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA512",
			keySizeInBytes: 16,
			tagSizeInBytes: 10,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s-%d-%d", tc.hkdfAlg, tc.tagAlg, tc.keySizeInBytes, tc.tagSizeInBytes), func(t *testing.T) {
			if tc.noFIPS && fips140.FIPSEnabled() {
				t.Skip("Skipping non-conforming test under FIPS mode.")
			}

			primitive, err := subtle.NewAESCTRHMAC(mainKey, tc.hkdfAlg, tc.keySizeInBytes, tc.tagAlg, tc.tagSizeInBytes, segmentSize, firstSegmentOffset)
			if err != nil {
				t.Fatalf("subtle.NewAESCTRHMAC err = %v, want nil", err)
			}
			ciphertextBuffer := &bytes.Buffer{}
			_, err = primitive.NewEncryptingWriter(ciphertextBuffer, []byte("associatedData"))
			if err != nil {
				t.Fatalf("primitive.NewEncryptingWriter err = %v, want nil", err)
			}
		})
	}
}

func TestAESCTRHMACWithInvalidParameters(t *testing.T) {
	const (
		keySizeInBytes     = 16
		tagSizeInBytes     = 12
		segmentSize        = 256
		firstSegmentOffset = 8
	)
	mainKey, err := hex.DecodeString(
		"000102030405060708090a0b0c0d0e0f00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		hkdfAlg        string
		tagAlg         string
		keySizeInBytes int
		tagSizeInBytes int
	}{
		// keySize too small
		{
			hkdfAlg:        "SHA1",
			tagAlg:         "SHA1",
			keySizeInBytes: 15,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA256",
			keySizeInBytes: 15,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA224",
			tagAlg:         "SHA224",
			keySizeInBytes: 15,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA384",
			tagAlg:         "SHA384",
			keySizeInBytes: 15,
			tagSizeInBytes: 10,
		},
		{
			hkdfAlg:        "SHA512",
			tagAlg:         "SHA512",
			keySizeInBytes: 15,
			tagSizeInBytes: 10,
		},
		// tagSize too small
		{
			hkdfAlg:        "SHA1",
			tagAlg:         "SHA1",
			keySizeInBytes: 16,
			tagSizeInBytes: 9,
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 9,
		},
		{
			hkdfAlg:        "SHA224",
			tagAlg:         "SHA224",
			keySizeInBytes: 16,
			tagSizeInBytes: 9,
		},
		{
			hkdfAlg:        "SHA384",
			tagAlg:         "SHA384",
			keySizeInBytes: 16,
			tagSizeInBytes: 9,
		},
		{
			hkdfAlg:        "SHA512",
			tagAlg:         "SHA512",
			keySizeInBytes: 16,
			tagSizeInBytes: 9,
		},
		// tagSize larger than digest size
		{
			hkdfAlg:        "SHA1",
			tagAlg:         "SHA1",
			keySizeInBytes: 16,
			tagSizeInBytes: 21,
		},
		{
			hkdfAlg:        "SHA256",
			tagAlg:         "SHA256",
			keySizeInBytes: 16,
			tagSizeInBytes: 33,
		},
		{
			hkdfAlg:        "SHA224",
			tagAlg:         "SHA224",
			keySizeInBytes: 16,
			tagSizeInBytes: 29,
		},
		{
			hkdfAlg:        "SHA384",
			tagAlg:         "SHA384",
			keySizeInBytes: 16,
			tagSizeInBytes: 49,
		},
		{
			hkdfAlg:        "SHA512",
			tagAlg:         "SHA512",
			keySizeInBytes: 16,
			tagSizeInBytes: 65,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s-%d-%d", tc.hkdfAlg, tc.tagAlg, tc.keySizeInBytes, tc.tagSizeInBytes), func(t *testing.T) {
			_, err := subtle.NewAESCTRHMAC(mainKey, tc.hkdfAlg, tc.keySizeInBytes, tc.tagAlg, tc.tagSizeInBytes, segmentSize, firstSegmentOffset)
			if err == nil {
				t.Error("subtle.NewAESCTRHMAC = nil, want error")
			}
		})
	}
}

func TestAESCTRHMACWithNegativeFirstSegmentOffsetFails(t *testing.T) {
	const (
		keySizeInBytes     = 16
		tagSizeInBytes     = 12
		segmentSize        = 256
		hkdfAlg            = "SHA256"
		tagAlg             = "SHA256"
		firstSegmentOffset = -1
	)
	mainKey, err := hex.DecodeString(
		"000102030405060708090a0b0c0d0e0f00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}

	_, err = subtle.NewAESCTRHMAC(mainKey, hkdfAlg, keySizeInBytes, tagAlg, tagSizeInBytes, segmentSize, firstSegmentOffset)
	if err == nil {
		t.Error("subtle.NewAESCTRHMAC() = nil, want error")
	}
}
