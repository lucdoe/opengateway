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
	"net/http"

	"github.com/lucdoe/open-gateway/gateway/internal/plugins/cors"
)

type CORSMiddleware struct {
	CORS cors.CORS
}

func NewCORSMiddleware(cors cors.CORS) *CORSMiddleware {
	return &CORSMiddleware{
		CORS: cors,
	}
}

func (cm *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cm.CORS.ValidateOrigin(r.Header.Get("Origin")) {
			http.Error(w, "CORS policy violation: Invalid Origin", http.StatusForbidden)
			return
		}

		if r.Method == "OPTIONS" {
			if !cm.CORS.ValidateMethod(r.Header.Get("Access-Control-Request-Method")) ||
				!cm.CORS.ValidateHeaders(r.Header.Get("Access-Control-Request-Headers")) {
				http.Error(w, "CORS policy violation: Invalid Method or Header", http.StatusForbidden)
				return
			}
			cm.setCORSHeaders(w, r.Header.Get("Origin"))
			w.WriteHeader(http.StatusOK)
			return
		}

		cm.setCORSHeaders(w, r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	})
}

func (cm *CORSMiddleware) setCORSHeaders(w http.ResponseWriter, origin string) {
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", cm.CORS.GetAllowedMethods())
	w.Header().Set("Access-Control-Allow-Headers", cm.CORS.GetAllowedHeaders())
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
