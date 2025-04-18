// Copyright 2019 Google LLC
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

package aead_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"testing"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/testing/fakekms"
	tinkpb "github.com/tink-crypto/tink-go/v2/proto/tink_go_proto"
)

func TestKMSEnvelopeWorksWithTinkKeyTemplatesAsDekTemplate(t *testing.T) {
	keyURI := "fake-kms://CM2b3_MDElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEIK75t5L-adlUwVhWvRuWUwYARABGM2b3_MDIAE"
	kekAEAD, err := fakekms.NewAEAD(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEAD(keyURI) err = %q, want nil", err)
	}
	kekAEADWithContext, err := fakekms.NewAEADWithContext(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEADWithContext(keyURI) err = %q, want nil", err)
	}
	plaintext := []byte("plaintext")
	associatedData := []byte("associatedData")
	invalidAssociatedData := []byte("invalidAssociatedData")

	var kmsEnvelopeAeadDekTestCases = []struct {
		name        string
		dekTemplate *tinkpb.KeyTemplate
	}{
		{
			name:        "AES128_GCM",
			dekTemplate: aead.AES128GCMKeyTemplate(),
		}, {
			name:        "AES256_GCM",
			dekTemplate: aead.AES256GCMKeyTemplate(),
		}, {
			name:        "AES256_GCM_NO_PREFIX",
			dekTemplate: aead.AES256GCMNoPrefixKeyTemplate(),
		}, {
			name:        "AES128_GCM_SIV",
			dekTemplate: aead.AES128GCMSIVKeyTemplate(),
		}, {
			name:        "AES256_GCM_SIV",
			dekTemplate: aead.AES256GCMSIVKeyTemplate(),
		}, {
			name:        "AES256_GCM_SIV_NO_PREFIX",
			dekTemplate: aead.AES256GCMSIVNoPrefixKeyTemplate(),
		}, {
			name:        "AES128_CTR_HMAC_SHA256",
			dekTemplate: aead.AES128CTRHMACSHA256KeyTemplate(),
		}, {
			name:        "AES256_CTR_HMAC_SHA256",
			dekTemplate: aead.AES256CTRHMACSHA256KeyTemplate(),
		}, {
			name:        "CHACHA20_POLY1305",
			dekTemplate: aead.ChaCha20Poly1305KeyTemplate(),
		}, {
			name:        "XCHACHA20_POLY1305",
			dekTemplate: aead.XChaCha20Poly1305KeyTemplate(),
		},
	}
	for _, tc := range kmsEnvelopeAeadDekTestCases {
		t.Run(tc.name, func(t *testing.T) {
			a := aead.NewKMSEnvelopeAEAD2(tc.dekTemplate, kekAEAD)
			ciphertext, err := a.Encrypt(plaintext, associatedData)
			if err != nil {
				t.Fatalf("a.Encrypt(plaintext, associatedData) err = %q, want nil", err)
			}
			gotPlaintext, err := a.Decrypt(ciphertext, associatedData)
			if err != nil {
				t.Fatalf("a.Decrypt(ciphertext, associatedData) err = %q, want nil", err)
			}
			if !bytes.Equal(gotPlaintext, plaintext) {
				t.Fatalf("got plaintext %q, want %q", gotPlaintext, plaintext)
			}
			if _, err = a.Decrypt(ciphertext, invalidAssociatedData); err == nil {
				t.Error("a.Decrypt(ciphertext, invalidAssociatedData) err = nil, want error")
			}

			ctx := context.Background()
			r, err := aead.NewKMSEnvelopeAEADWithContext(tc.dekTemplate, kekAEADWithContext)
			if err != nil {
				t.Error("a.DecryptWithContext(ctx, ciphertext, invalidAssociatedData) err = nil, want error")
			}
			ciphertext2, err := r.EncryptWithContext(ctx, plaintext, associatedData)
			if err != nil {
				t.Fatalf("a.EncryptWithContext(ctx, plaintext, associatedData) err = %q, want nil", err)
			}
			gotPlaintext2, err := r.DecryptWithContext(ctx, ciphertext2, associatedData)
			if err != nil {
				t.Fatalf("a.DecryptWithContext(ctx, ciphertext2, associatedData) err = %q, want nil", err)
			}
			if !bytes.Equal(gotPlaintext2, plaintext) {
				t.Fatalf("got plaintext %q, want %q", gotPlaintext, plaintext)
			}
			if _, err = r.DecryptWithContext(ctx, ciphertext2, invalidAssociatedData); err == nil {
				t.Error("a.DecryptWithContext(ctx, ciphertext2, invalidAssociatedData) err = nil, want error")
			}

			// check that DecryptWithContext is compatible with Decrypt
			gotPlaintext3, err := r.DecryptWithContext(ctx, ciphertext, associatedData)
			if err != nil {
				t.Fatalf("r.DecryptWithContext(ctx, ciphertext, associatedData) err = %q, want nil", err)
			}
			if !bytes.Equal(gotPlaintext3, plaintext) {
				t.Fatalf("got plaintext %q, want %q", gotPlaintext3, plaintext)
			}
			gotPlaintext4, err := a.Decrypt(ciphertext2, associatedData)
			if err != nil {
				t.Fatalf("a.Decrypt(ciphertext2, associatedData) err = %q, want nil", err)
			}
			if !bytes.Equal(gotPlaintext4, plaintext) {
				t.Fatalf("got plaintext %q, want %q", gotPlaintext4, plaintext)
			}
		})
	}
}

func TestKMSEnvelopeDecryptTestVector(t *testing.T) {
	keyURI := "fake-kms://CM2b3_MDElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEIK75t5L-adlUwVhWvRuWUwYARABGM2b3_MDIAE"
	plaintext := []byte("plaintext")
	associatedData := []byte("associatedData")
	// Generated by running a.Decrypt([]byte("plaintext"), []byte("associatedData"))
	ciphertextHex := "00000043013e77cdcd3ac4f38c7312f97a9b8c6e2b7b481ce3540006d24abb658938b52d6474c5fa569212c023e06229c0335f414244a32748778baed5e24a55b12b82920d1f06f17ef8a86eee808d2c9d2026fb45371cf60696eb77790524642c5f067a529793c251b068f8"
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		t.Fatalf("hex.DecodeString(ciphertextHex) err = %q, want nil", err)
	}

	// with NewKMSEnvelopeAEAD2.
	kekAEAD, err := fakekms.NewAEAD(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEAD(keyURI) err = %q, want nil", err)
	}
	a := aead.NewKMSEnvelopeAEAD2(aead.AES256GCMKeyTemplate(), kekAEAD)
	gotPlaintext, err := a.Decrypt(ciphertext, associatedData)
	if err != nil {
		t.Fatalf("a.Decrypt(ciphertext, associatedData) err = %q, want nil", err)
	}
	if !bytes.Equal(gotPlaintext, plaintext) {
		t.Fatalf("got plaintext %q, want %q", gotPlaintext, plaintext)
	}

	// with NewKMSEnvelopeAEADWithContext.
	ctx := context.Background()
	kekAEADWithContext, err := fakekms.NewAEADWithContext(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEADWithContext(keyURI) err = %q, want nil", err)
	}
	r, err := aead.NewKMSEnvelopeAEADWithContext(aead.AES256GCMKeyTemplate(), kekAEADWithContext)
	if err != nil {
		t.Fatalf("aead.NewKMSEnvelopeAEADWithContext() err = %q, want nil", err)
	}
	gotPlaintext2, err := r.DecryptWithContext(ctx, ciphertext, associatedData)
	if err != nil {
		t.Fatalf("r.DecryptWithContext(ciphertext, associatedData) err = %q, want nil", err)
	}
	if !bytes.Equal(gotPlaintext2, plaintext) {
		t.Fatalf("got plaintext %q, want %q", gotPlaintext, plaintext)
	}
}

func TestKMSEnvelopeWithKmsEnvelopeKeyTemplatesAsDekTemplate_fails(t *testing.T) {
	keyURI := "fake-kms://CM2b3_MDElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEIK75t5L-adlUwVhWvRuWUwYARABGM2b3_MDIAE"
	plaintext := []byte("plaintext")
	associatedData := []byte("associatedData")
	// Use a KmsEnvelopeAeadKeyTemplate as DEK template.
	envelopeDEKTemplate, err := aead.CreateKMSEnvelopeAEADKeyTemplate(keyURI, aead.AES128GCMKeyTemplate())
	if err != nil {
		t.Fatalf("aead.CreateKMSEnvelopAEADKeyTemplate() err = %q, want nil", err)
	}

	// NewKMSEnvelopeAEAD2 can't return an error. But it always fails when calling Encrypt.
	kekAEAD, err := fakekms.NewAEAD(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEAD(keyURI) err = %q, want nil", err)
	}
	a := aead.NewKMSEnvelopeAEAD2(envelopeDEKTemplate, kekAEAD)
	_, err = a.Encrypt(plaintext, associatedData)
	if err == nil {
		t.Error("a.Encrypt(plaintext, associatedData) err = nil, want error")
	}

	// NewKMSEnvelopeAEADWithContext returns an error.
	kekAEADWithContext, err := fakekms.NewAEADWithContext(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEADWithContext(keyURI) err = %q, want nil", err)
	}
	_, err = aead.NewKMSEnvelopeAEADWithContext(envelopeDEKTemplate, kekAEADWithContext)
	if err == nil {
		t.Error("NewKMSEnvelopeAEADWithContext() err = nil, want error")
	}
}

func TestKMSEnvelopeShortCiphertext(t *testing.T) {
	keyURI := "fake-kms://CM2b3_MDElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEIK75t5L-adlUwVhWvRuWUwYARABGM2b3_MDIAE"

	// with NewKMSEnvelopeAEAD2.
	kekAEAD, err := fakekms.NewAEAD(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEAD(keyURI) err = %q, want nil", err)
	}
	a := aead.NewKMSEnvelopeAEAD2(aead.AES256GCMKeyTemplate(), kekAEAD)
	if _, err = a.Decrypt([]byte{1}, nil); err == nil {
		t.Error("a.Decrypt([]byte{1}, nil) err = nil, want error")
	}

	// with NewKMSEnvelopeAEADWithContext.
	kekAEADWithContext, err := fakekms.NewAEADWithContext(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEADWithContext(keyURI) err = %q, want nil", err)
	}
	r, err := aead.NewKMSEnvelopeAEADWithContext(aead.AES256GCMKeyTemplate(), kekAEADWithContext)
	if err != nil {
		t.Fatalf("fakekms.NewKMSEnvelopeAEADWithContext() err = %q, want nil", err)
	}
	if _, err = r.DecryptWithContext(context.Background(), []byte{1}, nil); err == nil {
		t.Error("a.DecryptWithContext([]byte{1}, nil) err = nil, want error")
	}
}

func TestKMSEnvelopeDecryptHugeEncryptedDek(t *testing.T) {
	keyURI := "fake-kms://CM2b3_MDElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEIK75t5L-adlUwVhWvRuWUwYARABGM2b3_MDIAE"
	// A ciphertext with a huge encrypted DEK length
	ciphertext := []byte{0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88}

	// with NewKMSEnvelopeAEAD2.
	kekAEAD, err := fakekms.NewAEAD(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEAD(keyURI) err = %q, want nil", err)
	}
	a := aead.NewKMSEnvelopeAEAD2(aead.AES256GCMKeyTemplate(), kekAEAD)

	if _, err = a.Decrypt(ciphertext, nil); err == nil {
		t.Error("a.Decrypt([]byte{1}, nil) err = nil, want error")
	}

	// with NewKMSEnvelopeAEADWithContext.
	ctx := context.Background()
	kekAEADWithContext, err := fakekms.NewAEADWithContext(keyURI)
	if err != nil {
		t.Fatalf("fakekms.NewAEADWithContext(keyURI) err = %q, want nil", err)
	}
	r, err := aead.NewKMSEnvelopeAEADWithContext(aead.AES256GCMKeyTemplate(), kekAEADWithContext)
	if err != nil {
		t.Fatalf("fakekms.NewKMSEnvelopeAEADWithContext() err = %q, want nil", err)
	}
	if _, err = r.DecryptWithContext(ctx, ciphertext, nil); err == nil {
		t.Error("a.Decrypt([]byte{1}, nil) err = nil, want error")
	}
}

type invalidAEAD struct {
}

func (a *invalidAEAD) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	return []byte{}, nil
}

func (a *invalidAEAD) Decrypt(ciphertext, associatedData []byte) ([]byte, error) {
	return []byte{}, nil
}

func TestKMSEnvelopeEncryptWithInvalidAEADFails(t *testing.T) {
	invalidKEKAEAD := &invalidAEAD{}
	envAEADWithInvalidKEK := aead.NewKMSEnvelopeAEAD2(aead.AES256GCMKeyTemplate(), invalidKEKAEAD)

	if _, err := envAEADWithInvalidKEK.Encrypt([]byte("plaintext"), []byte("associatedData")); err == nil {
		t.Error("envAEADWithInvalidKEK.Encrypt(plaintext, associatedData) err = nil, want error")
	}
}
