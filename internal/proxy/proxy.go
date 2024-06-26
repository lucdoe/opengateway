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

package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ProxyService interface {
	ReverseProxy(target string, w http.ResponseWriter, r *http.Request) error
}

type Proxy struct{}

func NewProxyService() ProxyService {
	return &Proxy{}
}

func (p *Proxy) ReverseProxy(target string, w http.ResponseWriter, r *http.Request) error {
	parsedURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, "Failed to parse target URL", http.StatusInternalServerError)
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = parsedURL.Scheme
		req.URL.Host = parsedURL.Host
		req.URL.Path = parsedURL.Path
		req.Host = parsedURL.Host
	}

	proxy.ModifyResponse = func(response *http.Response) error {
		return nil
	}

	proxy.ServeHTTP(w, r)
	return nil
}
