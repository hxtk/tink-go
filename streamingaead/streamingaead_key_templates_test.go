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

package streamingaead_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/tink-crypto/tink-go/v2/internal/fips140"
	"github.com/tink-crypto/tink-go/v2/keyset"
	tinkpb "github.com/tink-crypto/tink-go/v2/proto/tink_go_proto"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
)

func TestKeyTemplates(t *testing.T) {
	var testCases = []struct {
		name     string
		template *tinkpb.KeyTemplate
		noFIPS   bool
	}{
		{
			name:     "AES128_GCM_HKDF_4KB",
			template: streamingaead.AES128GCMHKDF4KBKeyTemplate(),

			// Uses GCM with non-FIPS-conforming nonce-generation scheme.
			// See FIPS140-3 Annex C.H. for approved generation scehmes. Sections
			// 4 and 5 require validation under the Cryptographic Module Verification Program
			// and would therefore have to be implemented within the FIPS module boundary, which
			// is beyond the scope of Tink.
			noFIPS: true,
		},
		{
			name:     "AES128_GCM_HKDF_1MB",
			template: streamingaead.AES128GCMHKDF1MBKeyTemplate(),
			noFIPS:   true, // Non-conforming GCM Nonce
		},
		{
			name:     "AES256_GCM_HKDF_4KB",
			template: streamingaead.AES256GCMHKDF4KBKeyTemplate(),
			noFIPS:   true, // Non-conforming GCM Nonce
		}, {
			name:     "AES256_GCM_HKDF_1MB",
			template: streamingaead.AES256GCMHKDF1MBKeyTemplate(),
			noFIPS:   true, // Non-conforming GCM Nonce
		}, {
			name:     "AES128_CTR_HMAC_SHA256_4KB",
			template: streamingaead.AES128CTRHMACSHA256Segment4KBKeyTemplate(),
		},
		{
			name:     "AES128_CTR_HMAC_SHA256_1MB",
			template: streamingaead.AES128CTRHMACSHA256Segment1MBKeyTemplate(),
		},
		{
			name:     "AES256_CTR_HMAC_SHA256_4KB",
			template: streamingaead.AES256CTRHMACSHA256Segment4KBKeyTemplate(),
		},
		{
			name:     "AES256_CTR_HMAC_SHA256_1MB",
			template: streamingaead.AES256CTRHMACSHA256Segment1MBKeyTemplate(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.noFIPS && fips140.FIPSEnabled() {
				t.Skipf("Skipping %s under FIPS mode due to expected non-conformance.", tc.name)
			}
			handle, err := keyset.NewHandle(tc.template)
			if err != nil {
				t.Fatalf("keyset.NewHandle(template) failed: %v", err)
			}
			primitive, err := streamingaead.New(handle)
			if err != nil {
				t.Fatalf("aead.New(handle) failed: %v", err)
			}

			plaintext := []byte("some data to encrypt")
			aad := []byte("extra data to authenticate")
			buf := &bytes.Buffer{}
			w, err := primitive.NewEncryptingWriter(buf, aad)
			if err != nil {
				t.Fatalf("primitive.NewEncryptingWriter(buf, aad) failed: %v", err)
			}
			if _, err := w.Write(plaintext); err != nil {
				t.Fatalf("w.Write(plaintext) failed: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("w.Close() failed: %v", err)
			}

			r, err := primitive.NewDecryptingReader(buf, aad)
			if err != nil {
				t.Fatalf("primitive.NewDecryptingReader(buf, aad) failed: %v", err)
			}
			decrypted, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("io.ReadAll(r) failed: %v", err)
			}
			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("decrypted data doesn't match plaintext, got: %q, want: ''", decrypted)
			}
		})
	}
}
