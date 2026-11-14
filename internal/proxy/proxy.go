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
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// NewHandler returns a handler that forwards every request it receives to
// the upstream at target. The incoming path and query are kept and appended
// to target's own path, so a request for /api/people/7 sent to
// https://example.com goes to https://example.com/api/people/7.
//
// The handler is safe for concurrent use and meant to be built once per
// upstream and reused. If the upstream cannot be reached it answers
// 502 Bad Gateway. If the upstream has not started answering within timeout
// it answers 504 Gateway Timeout; a timeout of 0 waits without limit.
func NewHandler(target string, timeout time.Duration) (http.Handler, error) {
	upstream, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL %q: %w", target, err)
	}
	if (upstream.Scheme != "http" && upstream.Scheme != "https") || upstream.Host == "" {
		return nil, fmt.Errorf("invalid upstream URL %q: must be an absolute http or https URL", target)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = timeout

	return &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(upstream)
			r.SetXForwarded()
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy: %s %s to %s failed: %v", r.Method, r.URL.Path, upstream.Host, err)
			if isTimeout(err) {
				http.Error(w, "Gateway timeout", http.StatusGatewayTimeout)
				return
			}
			http.Error(w, "Bad gateway", http.StatusBadGateway)
		},
	}, nil
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
