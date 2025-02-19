// Copyright 2022 Google LLC
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

package jwt

import (
	"fmt"

	"github.com/tink-crypto/tink-go/v2/internal/internalapi"
	"github.com/tink-crypto/tink-go/v2/internal/internalregistry"
	"github.com/tink-crypto/tink-go/v2/internal/monitoringutil"
	"github.com/tink-crypto/tink-go/v2/internal/primitiveset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/monitoring"
	tinkpb "github.com/tink-crypto/tink-go/v2/proto/tink_go_proto"
)

// NewVerifier generates a new instance of the JWT Verifier primitive.
func NewVerifier(handle *keyset.Handle) (Verifier, error) {
	if handle == nil {
		return nil, fmt.Errorf("keyset handle can't be nil")
	}
	ps, err := keyset.Primitives[*verifierWithKID](handle, internalapi.Token{})
	if err != nil {
		return nil, fmt.Errorf("jwt_verifier_factory: cannot obtain primitive set: %v", err)
	}
	return newWrappedVerifier(ps)
}

// wrappedVerifier is a JWT Verifier implementation that uses the underlying primitive set for JWT Verifier.
type wrappedVerifier struct {
	ps     *primitiveset.PrimitiveSet[*verifierWithKID]
	logger monitoring.Logger
}

var _ Verifier = (*wrappedVerifier)(nil)

func createVerifierLogger(ps *primitiveset.PrimitiveSet[*verifierWithKID]) (monitoring.Logger, error) {
	// only keysets which contain annotations are monitored.
	if len(ps.Annotations) == 0 {
		return &monitoringutil.DoNothingLogger{}, nil
	}
	keysetInfo, err := monitoringutil.KeysetInfoFromPrimitiveSet(ps)
	if err != nil {
		return nil, err
	}
	return internalregistry.GetMonitoringClient().NewLogger(&monitoring.Context{
		KeysetInfo:  keysetInfo,
		Primitive:   "jwtverify",
		APIFunction: "verify",
	})
}

func newWrappedVerifier(ps *primitiveset.PrimitiveSet[*verifierWithKID]) (*wrappedVerifier, error) {
	for _, primitives := range ps.Entries {
		for _, p := range primitives {
			if p.PrefixType != tinkpb.OutputPrefixType_RAW && p.PrefixType != tinkpb.OutputPrefixType_TINK {
				return nil, fmt.Errorf("jwt_verifier_factory: invalid OutputPrefixType: %s", p.PrefixType)
			}
		}
	}
	logger, err := createVerifierLogger(ps)
	if err != nil {
		return nil, err
	}
	return &wrappedVerifier{
		ps:     ps,
		logger: logger,
	}, nil
}

func (w *wrappedVerifier) VerifyAndDecode(compact string, validator *Validator) (*VerifiedJWT, error) {
	var interestingErr error
	for _, s := range w.ps.Entries {
		for _, e := range s {
			verifiedJWT, err := e.Primitive.VerifyAndDecodeWithKID(compact, validator, keyID(e.KeyID, e.PrefixType))
			if err == nil {
				w.logger.Log(e.KeyID, 1)
				return verifiedJWT, nil
			}
			if err != errJwtVerification {
				// any error that is not the generic errJwtVerification is considered interesting
				interestingErr = err
			}
		}
	}
	w.logger.LogFailure()
	if interestingErr != nil {
		return nil, interestingErr
	}
	return nil, errJwtVerification
}
