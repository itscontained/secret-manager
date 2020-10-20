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

package log

import (
	"context"

	"github.com/go-logr/logr"
)

var (
	contextKey = &struct{}{}
)

// FromContext returns a logger with predefined values from a context.Context.
func FromContext(ctx context.Context, keysAndValues ...interface{}) logr.Logger {
	var log logr.Logger

	if ctx != nil {
		if lv := ctx.Value(contextKey); lv != nil {
			log = lv.(logr.Logger)
		}
	}

	return log.WithValues(keysAndValues...)
}

// IntoContext takes a context and sets the logger as one of its keys.
// Use FromContext function to retrieve the logger.
func IntoContext(ctx context.Context, log logr.Logger) context.Context {
	return context.WithValue(ctx, contextKey, log)
}
