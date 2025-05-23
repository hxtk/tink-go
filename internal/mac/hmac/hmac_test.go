// Copyright 2024 Google LLC
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

package hmac_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/tink-crypto/tink-go/v2/internal/mac/hmac"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

var key, _ = hex.DecodeString("000102030405060708090a0b0c0d0e0f")
var data = []byte("Hello")
var hmacTests = []struct {
	desc        string
	hashAlg     string
	tagSize     uint32
	key         []byte
	data        []byte
	expectedMac string
}{
	{
		desc:        "with SHA256 and 32 byte tag",
		hashAlg:     "SHA256",
		tagSize:     32,
		data:        data,
		key:         key,
		expectedMac: "e0ff02553d9a619661026c7aa1ddf59b7b44eac06a9908ff9e19961d481935d4",
	},
	{
		desc:    "with SHA512 and 64 byte tag",
		hashAlg: "SHA512",
		tagSize: 64,
		data:    data,
		key:     key,
		expectedMac: "481e10d823ba64c15b94537a3de3f253c16642451ac45124dd4dde120bf1e5c15" +
			"e55487d55ba72b43039f235226e7954cd5854b30abc4b5b53171a4177047c9b",
	},
	// empty data
	{
		desc:        "empty data",
		hashAlg:     "SHA256",
		tagSize:     32,
		data:        []byte{},
		key:         key,
		expectedMac: "07eff8b326b7798c9ccfcbdbe579489ac785a7995a04618b1a2813c26744777d",
	},
}

func TestHMACBasic(t *testing.T) {
	for _, test := range hmacTests {
		t.Run(test.desc, func(t *testing.T) {
			cipher, err := hmac.New(test.hashAlg, test.key, test.tagSize)
			if err != nil {
				t.Fatalf("hmac.New() err = %q, want nil", err)
			}
			mac, err := cipher.ComputeMAC(test.data)
			if err != nil {
				t.Fatalf("cipher.ComputeMAC() err = %q, want nil", err)
			}
			if hex.EncodeToString(mac) != test.expectedMac {
				t.Errorf("hex.EncodeToString(mac) = %q, want %q",
					hex.EncodeToString(mac), test.expectedMac)
			}
			if err := cipher.VerifyMAC(mac, test.data); err != nil {
				t.Errorf("cipher.VerifyMAC() err = %q, want nil", err)
			}
		})
	}
}

func TestNewHMACWithInvalidInput(t *testing.T) {
	// invalid hash algorithm
	_, err := hmac.New("MD5", random.GetRandomBytes(16), 32)
	if err == nil || !strings.Contains(err.Error(), "invalid hash algorithm") {
		t.Errorf("expect an error when hash algorithm is invalid")
	}
	// key too short
	_, err = hmac.New("SHA256", random.GetRandomBytes(1), 32)
	if err == nil || !strings.Contains(err.Error(), "key too short") {
		t.Errorf("expect an error when key is too short")
	}
	// tag too short
	_, err = hmac.New("SHA256", random.GetRandomBytes(16), 9)
	if err == nil || !strings.Contains(err.Error(), "tag size too small") {
		t.Errorf("expect an error when tag size is too small")
	}
	// tag too big
	_, err = hmac.New("SHA1", random.GetRandomBytes(16), 21)
	if err == nil || !strings.Contains(err.Error(), "tag size too big") {
		t.Errorf("expect an error when tag size is too big")
	}
	_, err = hmac.New("SHA256", random.GetRandomBytes(16), 33)
	if err == nil || !strings.Contains(err.Error(), "tag size too big") {
		t.Errorf("expect an error when tag size is too big")
	}
	_, err = hmac.New("SHA512", random.GetRandomBytes(16), 65)
	if err == nil || !strings.Contains(err.Error(), "tag size too big") {
		t.Errorf("expect an error when tag size is too big")
	}
}

func TestHMACWithNilHashFunc(t *testing.T) {
	cipher, err := hmac.New("SHA256", random.GetRandomBytes(32), 32)
	if err != nil {
		t.Fatalf("hmac.New() err = %v", err)
	}

	// Modify exported field.
	cipher.HashFunc = nil

	if _, err := cipher.ComputeMAC([]byte{}); err == nil {
		t.Errorf("cipher.ComputerMAC() err = nil, want not nil")
	}
}

func TestHMAComputeVerifyWithNilInput(t *testing.T) {
	cipher, err := hmac.New("SHA256", random.GetRandomBytes(16), 32)
	if err != nil {
		t.Errorf("unexpected error when creating new HMAC")
	}
	tag, err := cipher.ComputeMAC(nil)
	if err != nil {
		t.Errorf("cipher.ComputeMAC(nil) failed: %v", err)
	}
	if err := cipher.VerifyMAC(tag, nil); err != nil {
		t.Errorf("cipher.VerifyMAC(tag, nil) failed: %v", err)
	}
}

func TestVerifyMACWithInvalidInput(t *testing.T) {
	cipher, err := hmac.New("SHA256", random.GetRandomBytes(16), 32)
	if err != nil {
		t.Errorf("unexpected error when creating new HMAC")
	}
	if err := cipher.VerifyMAC(nil, []byte{1}); err == nil {
		t.Errorf("expect an error when mac is nil")
	}
	if err := cipher.VerifyMAC([]byte{1}, nil); err == nil {
		t.Errorf("expect an error when data is nil")
	}
	if err := cipher.VerifyMAC(nil, nil); err == nil {
		t.Errorf("cipher.VerifyMAC(nil, nil) succeeded unexpectedly")
	}
}

func TestHMACModification(t *testing.T) {
	for _, test := range hmacTests {
		t.Run(test.desc, func(t *testing.T) {
			cipher, err := hmac.New(test.hashAlg, test.key, test.tagSize)
			if err != nil {
				t.Fatalf("hmac.New() err = %q, want nil", err)
			}
			mac, err := cipher.ComputeMAC(test.data)
			if err != nil {
				t.Fatalf("cipher.ComputeMAC() err = %q, want nil", err)
			}
			for i := 0; i < len(mac); i++ {
				tmp := mac[i]
				for j := 0; j < 8; j++ {
					mac[i] ^= 1 << uint8(j)
					err := cipher.VerifyMAC(mac, test.data)
					if err == nil {
						t.Errorf("cipher.VerifyMAC() of valid mac modified at position (%d, %d) is nil, want error", i, j)
					}
					mac[i] = tmp
				}
			}
		})
	}
}

func TestHMACTruncation(t *testing.T) {
	for i, test := range hmacTests {
		t.Run(test.desc, func(t *testing.T) {
			cipher, err := hmac.New(test.hashAlg, test.key, test.tagSize)
			if err != nil {
				t.Fatalf("hmac.New() err = %q, want nil", err)
			}
			mac, err := cipher.ComputeMAC(test.data)
			if err != nil {
				t.Fatalf("cipher.ComputeMAC() err = %q, want nil", err)
			}
			for truncatedLen := 1; i < len(mac); i++ {
				truncatedMAC := mac[:truncatedLen]
				err := cipher.VerifyMAC(truncatedMAC, test.data)
				if err == nil {
					t.Errorf("cipher.VerifyMAC() of a valid mac truncated to %d bytes is nil, want error",
						truncatedLen)
				}
			}
		})
	}
}
