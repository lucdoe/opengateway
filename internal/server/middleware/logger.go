// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"fmt"
	"net/http"

	"github.com/lucdoe/opengateway/internal/plugins/logger"
)

type LoggingMiddleware struct {
	Logger logger.Logger
}

func NewLoggingMiddleware(l logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		Logger: l,
	}
}

func (lm *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		details := fmt.Sprintf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		go lm.Logger.Info("Request", details)
		next.ServeHTTP(w, r)
	})
}
