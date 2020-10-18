/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schema

import (
	"encoding/json"
	"fmt"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/store"
)

var builder map[string]store.Factory

func init() {
	builder = make(map[string]store.Factory)
}

// Register a store backend type. Register panics if a
// backend with the same store is already registered
func Register(name string, s store.Factory) {
	_, exists := builder[name]
	if exists {
		panic(fmt.Sprintf("Store %q already registered", name))
	}

	builder[name] = s
}

// ForceRegister adds to store schema, overwriting a store if
// already registered. Should only be used for testing
func ForceRegister(name string, s store.Factory) {
	builder[name] = s
}

func GetStoreByName(name string) (store.Factory, bool) {
	f, ok := builder[name]
	return f, ok
}

func GetStore(store smv1alpha1.GenericStore) (store.Factory, error) {
	storeSpec := store.GetSpec()
	storeBytes, err := json.Marshal(storeSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal store spec: %w", err)
	}

	storeMap := make(map[string]interface{})
	err = json.Unmarshal(storeBytes, &storeMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal store spec: %w", err)
	}

	if len(storeMap) != 1 {
		return nil, fmt.Errorf("secret stores must only have exactly one backend specified, found %d for %s", len(storeMap), store.GetName())
	}

	for k := range storeMap {
		f, ok := builder[k]
		if !ok {
			return nil, fmt.Errorf("failed to find registered store backend for type: %s, name: %s", k, store.GetName())
		}
		return f, nil
	}

	return nil, fmt.Errorf("failed to find registered store backend for name: %s", store.GetName())
}
