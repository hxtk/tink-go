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

package daead

import (
	"fmt"

	"github.com/tink-crypto/tink-go/v2/core/cryptofmt"
	"github.com/tink-crypto/tink-go/v2/internal/internalapi"
	"github.com/tink-crypto/tink-go/v2/internal/internalregistry"
	"github.com/tink-crypto/tink-go/v2/internal/monitoringutil"
	"github.com/tink-crypto/tink-go/v2/internal/primitiveset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/monitoring"
	"github.com/tink-crypto/tink-go/v2/tink"
)

// New returns a DeterministicAEAD primitive from the given keyset handle.
func New(handle *keyset.Handle) (tink.DeterministicAEAD, error) {
	ps, err := keyset.Primitives[tink.DeterministicAEAD](handle, internalapi.Token{})
	if err != nil {
		return nil, fmt.Errorf("daead_factory: cannot obtain primitive set: %s", err)
	}
	return newWrappedDeterministicAEAD(ps)
}

// wrappedDeterministicAEAD is a DeterministicAEAD implementation that uses an underlying primitive set
// for deterministic encryption and decryption.
type wrappedDeterministicAEAD struct {
	ps        *primitiveset.PrimitiveSet[tink.DeterministicAEAD]
	encLogger monitoring.Logger
	decLogger monitoring.Logger
}

// Asserts that wrappedDeterministicAEAD implements the DeterministicAEAD interface.
var _ tink.DeterministicAEAD = (*wrappedDeterministicAEAD)(nil)

func newWrappedDeterministicAEAD(ps *primitiveset.PrimitiveSet[tink.DeterministicAEAD]) (*wrappedDeterministicAEAD, error) {
	encLogger, decLogger, err := createLoggers(ps)
	if err != nil {
		return nil, err
	}
	return &wrappedDeterministicAEAD{
		ps:        ps,
		encLogger: encLogger,
		decLogger: decLogger,
	}, nil
}

func createLoggers(ps *primitiveset.PrimitiveSet[tink.DeterministicAEAD]) (monitoring.Logger, monitoring.Logger, error) {
	if len(ps.Annotations) == 0 {
		return &monitoringutil.DoNothingLogger{}, &monitoringutil.DoNothingLogger{}, nil
	}
	client := internalregistry.GetMonitoringClient()
	keysetInfo, err := monitoringutil.KeysetInfoFromPrimitiveSet(ps)
	if err != nil {
		return nil, nil, err
	}
	encLogger, err := client.NewLogger(&monitoring.Context{
		Primitive:   "daead",
		APIFunction: "encrypt",
		KeysetInfo:  keysetInfo,
	})
	if err != nil {
		return nil, nil, err
	}
	decLogger, err := client.NewLogger(&monitoring.Context{
		Primitive:   "daead",
		APIFunction: "decrypt",
		KeysetInfo:  keysetInfo,
	})
	if err != nil {
		return nil, nil, err
	}
	return encLogger, decLogger, nil
}

// EncryptDeterministically deterministically encrypts plaintext with additionalData as additional authenticated data.
// It returns the concatenation of the primary's identifier and the ciphertext.
func (d *wrappedDeterministicAEAD) EncryptDeterministically(pt, aad []byte) ([]byte, error) {
	primary := d.ps.Primary
	p, ok := (primary.Primitive).(tink.DeterministicAEAD)
	if !ok {
		return nil, fmt.Errorf("daead_factory: not a DeterministicAEAD primitive")
	}

	ct, err := p.EncryptDeterministically(pt, aad)
	if err != nil {
		d.encLogger.LogFailure()
		return nil, err
	}
	d.encLogger.Log(primary.KeyID, len(pt))
	if len(primary.Prefix) == 0 {
		return ct, nil
	}
	output := make([]byte, 0, len(primary.Prefix)+len(ct))
	output = append(output, primary.Prefix...)
	output = append(output, ct...)
	return output, nil
}

// DecryptDeterministically deterministically decrypts ciphertext with additionalData as
// additional authenticated data. It returns the corresponding plaintext if the
// ciphertext is authenticated.
func (d *wrappedDeterministicAEAD) DecryptDeterministically(ct, aad []byte) ([]byte, error) {
	// try non-raw keys
	prefixSize := cryptofmt.NonRawPrefixSize
	if len(ct) > prefixSize {
		prefix := ct[:prefixSize]
		ctNoPrefix := ct[prefixSize:]
		entries, err := d.ps.EntriesForPrefix(string(prefix))
		if err == nil {
			for i := 0; i < len(entries); i++ {
				pt, err := entries[i].Primitive.DecryptDeterministically(ctNoPrefix, aad)
				if err == nil {
					d.decLogger.Log(entries[i].KeyID, len(ctNoPrefix))
					return pt, nil
				}
			}
		}
	}

	// try raw keys
	entries, err := d.ps.RawEntries()
	if err == nil {
		for i := 0; i < len(entries); i++ {
			pt, err := entries[i].Primitive.DecryptDeterministically(ct, aad)
			if err == nil {
				d.decLogger.Log(entries[i].KeyID, len(ct))
				return pt, nil
			}
		}
	}
	// nothing worked
	d.decLogger.LogFailure()
	return nil, fmt.Errorf("daead_factory: decryption failed")
}
