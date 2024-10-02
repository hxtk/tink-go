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

// Package configuration provides internal implementation of Configurations.
package configuration

import (
	"fmt"
	"reflect"

	"github.com/tink-crypto/tink-go/v2/internal/internalapi"
	"github.com/tink-crypto/tink-go/v2/key"
)

// Configuration keeps a collection of functions that create a primitive from
// [key.Key].
//
// This is an internal API.
type Configuration struct {
	primitiveCreators map[reflect.Type]primitiveConstructor
}

type primitiveConstructor func(key key.Key) (any, error)

// PrimitiveFromKey creates a primitive from the given [key.Key]. Returns an
// error if there is no primitiveConstructor registered for the given key.
//
// This is an internal API.
func (c *Configuration) PrimitiveFromKey(k key.Key, _ internalapi.Token) (any, error) {
	keyType := reflect.TypeOf(k)
	creator, ok := c.primitiveCreators[keyType]
	if !ok {
		return nil, fmt.Errorf("PrimitiveFromKey: no primitive creator from key %v registered", keyType)
	}
	return creator(k)
}

// RegisterPrimitiveConstructor registers a primitiveConstructor for the keyType.
// Not thread-safe.
//
// Returns an error if a primitiveConstructor for the keyType already
// registered (no matter whether it's the same object or different, since
// constructors are of type [Func] and they are never considered equal in Go
// unless they are nil).
//
// This is an internal API.
func (c *Configuration) RegisterPrimitiveConstructor(keyType reflect.Type, constructor primitiveConstructor, _ internalapi.Token) error {
	if _, ok := c.primitiveCreators[keyType]; ok {
		return fmt.Errorf("RegisterPrimitiveConstructor: attempt to register a different primitive constructor for the same key type %v", keyType)
	}
	c.primitiveCreators[keyType] = constructor
	return nil
}

// New creates an empty Configuration.
func New() (*Configuration, error) {
	return &Configuration{map[reflect.Type]primitiveConstructor{}}, nil
}
