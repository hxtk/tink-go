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

// Package mac provides implementations of the MAC primitive.
//
// MAC computes a tag for a given message that can be used to authenticate a
// message.  MAC protects data integrity as well as provides for authenticity
// of the message.
package mac

import (
	"fmt"

	"github.com/tink-crypto/tink-go/v2/core/registry"
	"github.com/tink-crypto/tink-go/v2/internal/internalregistry"
)

func init() {
	if err := registry.RegisterKeyManager(new(hmacKeyManager)); err != nil {
		panic(fmt.Sprintf("mac.init() failed: %v", err))
	}
	if err := internalregistry.AllowKeyDerivation(hmacTypeURL); err != nil {
		panic(fmt.Sprintf("mac.init() failed: %v", err))
	}
	if err := registry.RegisterKeyManager(new(aescmacKeyManager)); err != nil {
		panic(fmt.Sprintf("mac.init() failed: %v", err))
	}
}
