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
	"sync"

	smv1alpha1 "github.com/itscontained/secret-manager/pkg/apis/secretmanager/v1alpha1"
	"github.com/itscontained/secret-manager/pkg/store"
)

var builder map[string]store.Factory
var buildlock sync.RWMutex

func init() {
	builder = make(map[string]store.Factory)
}

// Register a store backend type. Register panics if a
// backend with the same store is already registered
func Register(s store.Factory, storeSpec *smv1alpha1.SecretStoreSpec) {
	storeName, err := getStoreBackend(storeSpec)
	if err != nil {
		panic(fmt.Sprintf("store error registering schema: %s", err.Error()))
	}

	buildlock.Lock()
	defer buildlock.Unlock()
	_, exists := builder[storeName]
	if exists {
		panic(fmt.Sprintf("store %q already registered", storeName))
	}

	builder[storeName] = s
}

// ForceRegister adds to store schema, overwriting a store if
// already registered. Should only be used for testing
func ForceRegister(s store.Factory, storeSpec *smv1alpha1.SecretStoreSpec) {
	storeName, err := getStoreBackend(storeSpec)
	if err != nil {
		panic(fmt.Sprintf("store error registering schema: %s", err.Error()))
	}

	buildlock.Lock()
	builder[storeName] = s
	buildlock.Unlock()
}

func GetStoreByName(name string) (store.Factory, bool) {
	buildlock.RLock()
	f, ok := builder[name]
	buildlock.RUnlock()
	return f, ok
}

func GetStore(store smv1alpha1.GenericStore) (store.Factory, error) {
	storeSpec := store.GetSpec()
	storeName, err := getStoreBackend(storeSpec)
	if err != nil {
		return nil, fmt.Errorf("store error for %s: %w", store.GetName(), err)
	}

	buildlock.RLock()
	f, ok := builder[storeName]
	buildlock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("failed to find registered store backend for type: %s, name: %s", storeName, store.GetName())
	}
	return f, nil
}

func getStoreBackend(storeSpec *smv1alpha1.SecretStoreSpec) (string, error) {
	storeBytes, err := json.Marshal(storeSpec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal store spec: %w", err)
	}

	storeMap := make(map[string]interface{})
	err = json.Unmarshal(storeBytes, &storeMap)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal store spec: %w", err)
	}

	if len(storeMap) != 1 {
		return "", fmt.Errorf("secret stores must only have exactly one backend specified, found %d", len(storeMap))
	}

	for k := range storeMap {
		return k, nil
	}

	return "", fmt.Errorf("failed to find registered store backend")
}
