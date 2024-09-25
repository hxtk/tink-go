// Copyright 2023 Google LLC
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

// Package registryconfig is a transitioning stepping stone used by the
// keyset handle in cases where a configuration is not provided by the user,
// so it needs to resort to using the old global registry methods.
package registryconfig

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/tink-crypto/tink-go/v2/core/registry"
	"github.com/tink-crypto/tink-go/v2/internal/internalapi"
	"github.com/tink-crypto/tink-go/v2/internal/protoserialization"
	"github.com/tink-crypto/tink-go/v2/key"
)

var (
	primitiveConstructorsMu sync.RWMutex
	primitiveConstructors   = make(map[reflect.Type]primitiveConstructor)
)

type primitiveConstructor func(key key.Key) (any, error)

// RegistryConfig is an internal way for the keyset handle to access the
// old global Registry through the new Configuration interface.
type RegistryConfig struct{}

// PrimitiveFromKey constructs a primitive from a [key.Key] using the registry.
func (c *RegistryConfig) PrimitiveFromKey(key key.Key, _ internalapi.Token) (any, error) {
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}
	constructor, found := primitiveConstructors[reflect.TypeOf(key)]
	if !found {
		// Fallback to using the key manager.
		keySerialization, err := protoserialization.SerializeKey(key)
		if err != nil {
			return nil, err
		}
		return registry.PrimitiveFromKeyData(keySerialization.KeyData())
	}
	return constructor(key)
}

// RegisterKeyManager registers a provided [registry.KeyManager] by forwarding
// it directly to the Registry.
func (c *RegistryConfig) RegisterKeyManager(km registry.KeyManager, _ internalapi.Token) error {
	return registry.RegisterKeyManager(km)
}

// RegisterPrimitiveConstructor registers a function that constructs primitives
// from a given [key.Key] to the global registry.
func RegisterPrimitiveConstructor[K key.Key](constructor primitiveConstructor) error {
	keyType := reflect.TypeFor[K]()
	primitiveConstructorsMu.Lock()
	defer primitiveConstructorsMu.Unlock()
	if existingCreator, found := primitiveConstructors[keyType]; found && reflect.ValueOf(existingCreator).Pointer() != reflect.ValueOf(constructor).Pointer() {
		return fmt.Errorf("a different constructor already registered for %v", keyType)
	}
	primitiveConstructors[keyType] = constructor
	return nil
}

// ClearPrimitiveConstructors clears the registry of primitive constructors.
//
// This function is intended to be used in tests only.
func ClearPrimitiveConstructors() {
	primitiveConstructorsMu.Lock()
	defer primitiveConstructorsMu.Unlock()
	primitiveConstructors = make(map[reflect.Type]primitiveConstructor)
}
