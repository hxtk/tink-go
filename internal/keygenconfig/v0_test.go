// Copyright 2025 Google LLC
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

package keygenconfig_test

import (
	"fmt"
	"testing"

	"github.com/tink-crypto/tink-go/v2/aead/aesctrhmac"
	"github.com/tink-crypto/tink-go/v2/aead/aesgcm"
	"github.com/tink-crypto/tink-go/v2/aead/aesgcmsiv"
	"github.com/tink-crypto/tink-go/v2/aead/chacha20poly1305"
	"github.com/tink-crypto/tink-go/v2/aead/xaesgcm"
	"github.com/tink-crypto/tink-go/v2/aead/xchacha20poly1305"
	"github.com/tink-crypto/tink-go/v2/internal/keygenconfig"
	"github.com/tink-crypto/tink-go/v2/key"
)

func mustCreateAESGCMParams(t *testing.T, variant aesgcm.Variant) *aesgcm.Parameters {
	t.Helper()
	params, err := aesgcm.NewParameters(aesgcm.ParametersOpts{
		KeySizeInBytes: 32,
		IVSizeInBytes:  12,
		TagSizeInBytes: 16,
		Variant:        variant,
	})
	if err != nil {
		t.Fatalf("aesgcm.NewParameters() err = %v, want nil", err)
	}
	return params
}

func mustCreateAESGCMSIVParams(t *testing.T, variant aesgcmsiv.Variant) *aesgcmsiv.Parameters {
	t.Helper()
	params, err := aesgcmsiv.NewParameters(32, variant)
	if err != nil {
		t.Fatalf("aesgcmsiv.NewParameters() err = %v, want nil", err)
	}
	return params
}

func mustCreateAESCTRHMACParams(t *testing.T, variant aesctrhmac.Variant) *aesctrhmac.Parameters {
	t.Helper()
	params, err := aesctrhmac.NewParameters(aesctrhmac.ParametersOpts{
		AESKeySizeInBytes:  32,
		HMACKeySizeInBytes: 32,
		IVSizeInBytes:      12,
		TagSizeInBytes:     16,
		HashType:           aesctrhmac.SHA256,
		Variant:            variant,
	})
	if err != nil {
		t.Fatalf("aesctrhmac.NewParameters() err = %v, want nil", err)
	}
	return params
}

func mustCreateChaCha20Poly1305Params(t *testing.T, variant chacha20poly1305.Variant) *chacha20poly1305.Parameters {
	t.Helper()
	params, err := chacha20poly1305.NewParameters(variant)
	if err != nil {
		t.Fatalf("chacha20poly1305.NewParameters() err = %v, want nil", err)
	}
	return params
}

func mustCreateXAESGCMParams(t *testing.T, variant xaesgcm.Variant) *xaesgcm.Parameters {
	t.Helper()
	params, err := xaesgcm.NewParameters(variant, 12)
	if err != nil {
		t.Fatalf("xaesgcm.NewParameters() err = %v, want nil", err)
	}
	return params
}

func mustCreateXChaCha20Poly1305Params(t *testing.T, variant xchacha20poly1305.Variant) *xchacha20poly1305.Parameters {
	t.Helper()
	params, err := xchacha20poly1305.NewParameters(variant)
	if err != nil {
		t.Fatalf("xchacha20poly1305.NewParameters() err = %v, want nil", err)
	}
	return params
}

func tryCast[T any](k key.Key) error {
	if _, ok := k.(T); !ok {
		return fmt.Errorf("key is of type %T; want %T", k, (*T)(nil))
	}
	return nil
}

func TestV0(t *testing.T) {
	config := keygenconfig.V0()
	for _, tc := range []struct {
		name          string
		p             key.Parameters
		idRequirement uint32
		tryCast       func(key.Key) error
	}{
		{
			name:          "AES-GCM-TINK",
			p:             mustCreateAESGCMParams(t, aesgcm.VariantTink),
			idRequirement: 123,
			tryCast:       tryCast[*aesgcm.Key],
		},
		{
			name:          "AES-GCM-NO_PREFIX",
			p:             mustCreateAESGCMParams(t, aesgcm.VariantNoPrefix),
			idRequirement: 0,
			tryCast:       tryCast[*aesgcm.Key],
		},
		{
			name:          "AES-CTR-HMAC-TINK",
			p:             mustCreateAESCTRHMACParams(t, aesctrhmac.VariantTink),
			idRequirement: 123,
			tryCast:       tryCast[*aesctrhmac.Key],
		},
		{
			name:          "AES-CTR-HMAC-NO_PREFIX",
			p:             mustCreateAESCTRHMACParams(t, aesctrhmac.VariantNoPrefix),
			idRequirement: 0,
			tryCast:       tryCast[*aesctrhmac.Key],
		},
		{
			name:          "AES-GCM-SIV-TINK",
			p:             mustCreateAESGCMSIVParams(t, aesgcmsiv.VariantTink),
			idRequirement: 123,
			tryCast:       tryCast[*aesgcmsiv.Key],
		},
		{
			name:          "AES-GCM-SIV-NO_PREFIX",
			p:             mustCreateAESGCMSIVParams(t, aesgcmsiv.VariantNoPrefix),
			idRequirement: 0,
			tryCast:       tryCast[*aesgcmsiv.Key],
		},
		{
			name:          "ChaCha20-Poly1305-TINK",
			p:             mustCreateChaCha20Poly1305Params(t, chacha20poly1305.VariantTink),
			idRequirement: 123,
			tryCast:       tryCast[*chacha20poly1305.Key],
		},
		{
			name:          "ChaCha20-Poly1305-NO_PREFIX",
			p:             mustCreateChaCha20Poly1305Params(t, chacha20poly1305.VariantNoPrefix),
			idRequirement: 0,
			tryCast:       tryCast[*chacha20poly1305.Key],
		},
		{
			name:          "XAES-GCM-TINK",
			p:             mustCreateXAESGCMParams(t, xaesgcm.VariantTink),
			idRequirement: 123,
			tryCast:       tryCast[*xaesgcm.Key],
		},
		{
			name:          "XAES-GCM-NO_PREFIX",
			p:             mustCreateXAESGCMParams(t, xaesgcm.VariantNoPrefix),
			idRequirement: 0,
			tryCast:       tryCast[*xaesgcm.Key],
		},
		{
			name:          "XChaCha20Poly1305-TINK",
			p:             mustCreateXChaCha20Poly1305Params(t, xchacha20poly1305.VariantTink),
			idRequirement: 123,
			tryCast:       tryCast[*xchacha20poly1305.Key],
		},
		{
			name:          "XChaCha20Poly1305-NO_PREFIX",
			p:             mustCreateXChaCha20Poly1305Params(t, xchacha20poly1305.VariantNoPrefix),
			idRequirement: 0,
			tryCast:       tryCast[*xchacha20poly1305.Key],
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, err := config.CreateKey(tc.p, tc.idRequirement)
			if err != nil {
				t.Fatalf("config.CreateKey(%v, %v) err = %v, want nil", tc.p, tc.idRequirement, err)
			}
			if err := tc.tryCast(key); err != nil {
				t.Errorf("tc.tryCast(key) = %v, want nil", err)
			}
		})
	}
}
