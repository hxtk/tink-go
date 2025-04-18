// Copyright 2018 Google LLC
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

package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
	"github.com/tink-crypto/tink-go/v2/internal/internalapi"
	"github.com/tink-crypto/tink-go/v2/internal/protoserialization"
	"github.com/tink-crypto/tink-go/v2/keyset"
	ed25519pb "github.com/tink-crypto/tink-go/v2/proto/ed25519_go_proto"
	tinkpb "github.com/tink-crypto/tink-go/v2/proto/tink_go_proto"
)

const (
	signerKeyVersion = 0
	signerTypeURL    = "type.googleapis.com/google.crypto.tink.Ed25519PrivateKey"
)

// common errors
var errInvalidSignKey = errors.New("invalid key")
var errInvalidSignKeyFormat = errors.New("invalid key format")

// signerKeyManager is an implementation of KeyManager interface.
// It generates new [ed25519pb.Ed25519PrivateKey] and produces new instances of
// [subtle.ED25519Signer].
type signerKeyManager struct{}

// Primitive creates a [subtle.ED25519Signer] instance for the given serialized
// [ed25519pb.Ed25519PrivateKey] proto.
func (km *signerKeyManager) Primitive(serializedKey []byte) (any, error) {
	keySerialization, err := protoserialization.NewKeySerialization(&tinkpb.KeyData{
		TypeUrl:         signerTypeURL,
		Value:           serializedKey,
		KeyMaterialType: tinkpb.KeyData_ASYMMETRIC_PRIVATE,
	}, tinkpb.OutputPrefixType_RAW, 0)
	if err != nil {
		return nil, err
	}
	key, err := protoserialization.ParseKey(keySerialization)
	if err != nil {
		return nil, err
	}
	signerKey, ok := key.(*PrivateKey)
	if !ok {
		return nil, fmt.Errorf("ed25519_signer_key_manager: invalid key type: got %T, want %T", key, (*PrivateKey)(nil))
	}
	return NewSigner(signerKey, internalapi.Token{})
}

// NewKey creates a new [ed25519pb.Ed25519PrivateKey] according to
// the given serialized [ed25519pb.Ed25519KeyFormat].
func (km *signerKeyManager) NewKey(serializedKeyFormat []byte) (proto.Message, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("cannot generate ED25519 key: %s", err)
	}
	return &ed25519pb.Ed25519PrivateKey{
		Version:  signerKeyVersion,
		KeyValue: priv.Seed(),
		PublicKey: &ed25519pb.Ed25519PublicKey{
			Version:  signerKeyVersion,
			KeyValue: pub,
		},
	}, nil
}

// NewKeyData creates a new KeyData according to specification in  the given
// serialized [ed25519pb.Ed25519KeyFormat]. It should be used solely by the key
// management API.
func (km *signerKeyManager) NewKeyData(serializedKeyFormat []byte) (*tinkpb.KeyData, error) {
	key, err := km.NewKey(serializedKeyFormat)
	if err != nil {
		return nil, err
	}
	serializedKey, err := proto.Marshal(key)
	if err != nil {
		return nil, errInvalidSignKeyFormat
	}
	return &tinkpb.KeyData{
		TypeUrl:         signerTypeURL,
		Value:           serializedKey,
		KeyMaterialType: km.KeyMaterialType(),
	}, nil
}

// PublicKeyData extracts the public key data from the private key.
func (km *signerKeyManager) PublicKeyData(serializedPrivKey []byte) (*tinkpb.KeyData, error) {
	privKey := new(ed25519pb.Ed25519PrivateKey)
	if err := proto.Unmarshal(serializedPrivKey, privKey); err != nil {
		return nil, errInvalidSignKey
	}
	serializedPubKey, err := proto.Marshal(privKey.PublicKey)
	if err != nil {
		return nil, errInvalidSignKey
	}
	return &tinkpb.KeyData{
		TypeUrl:         verifierTypeURL,
		Value:           serializedPubKey,
		KeyMaterialType: tinkpb.KeyData_ASYMMETRIC_PUBLIC,
	}, nil
}

// DoesSupport indicates if this key manager supports the given key type.
func (km *signerKeyManager) DoesSupport(typeURL string) bool { return typeURL == signerTypeURL }

// TypeURL returns the key type of keys managed by this key manager.
func (km *signerKeyManager) TypeURL() string { return signerTypeURL }

// KeyMaterialType returns the key material type of this key manager.
func (km *signerKeyManager) KeyMaterialType() tinkpb.KeyData_KeyMaterialType {
	return tinkpb.KeyData_ASYMMETRIC_PRIVATE
}

// DeriveKey derives a new key from serializedKeyFormat and pseudorandomness.
// Unlike NewKey, DeriveKey validates serializedKeyFormat's version.
func (km *signerKeyManager) DeriveKey(serializedKeyFormat []byte, pseudorandomness io.Reader) (proto.Message, error) {
	keyFormat := new(ed25519pb.Ed25519KeyFormat)
	if err := proto.Unmarshal(serializedKeyFormat, keyFormat); err != nil {
		return nil, err
	}
	err := keyset.ValidateKeyVersion(keyFormat.Version, signerKeyVersion)
	if err != nil {
		return nil, err
	}
	pub, priv, err := ed25519.GenerateKey(pseudorandomness)
	if err != nil {
		return nil, err
	}
	return &ed25519pb.Ed25519PrivateKey{
		Version:  signerKeyVersion,
		KeyValue: priv.Seed(),
		PublicKey: &ed25519pb.Ed25519PublicKey{
			Version:  signerKeyVersion,
			KeyValue: pub,
		},
	}, nil
}

// validateKey validates the given [ed25519pb.Ed25519PrivateKey].
func (km *signerKeyManager) validateKey(key *ed25519pb.Ed25519PrivateKey) error {
	if err := keyset.ValidateKeyVersion(key.Version, signerKeyVersion); err != nil {
		return fmt.Errorf("ed25519_signer_key_manager: invalid key: %s", err)
	}
	if len(key.KeyValue) != ed25519.SeedSize {
		return fmt.Errorf("ed25519_signer_key_manager: invalid key length, got %d", len(key.KeyValue))
	}
	return nil
}
