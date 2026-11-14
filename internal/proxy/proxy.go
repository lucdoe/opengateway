// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/lucdoe/opengateway/internal/httpx"
	"github.com/lucdoe/opengateway/internal/requestid"
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
	// The default keeps 2 idle connections per host, so under concurrent load
	// almost every request dials anew and the closed connections exhaust the
	// local ports. Each service has its own transport, so this is per service.
	transport.MaxIdleConns = 0
	transport.MaxIdleConnsPerHost = MaxIdleConnsPerUpstream

	return &httputil.ReverseProxy{
		Transport:  transport,
		BufferPool: buffers,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(upstream)
			r.SetXForwarded()
		},
		// The gateway already set the request ID on the response. An upstream
		// that echoes it would add a second value.
		ModifyResponse: func(res *http.Response) error {
			res.Header.Del(requestid.Header)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy: %s %s to %s failed: %v", r.Method, r.URL.Path, upstream.Host, err)
			if isTimeout(err) {
				httpx.Error(w, http.StatusGatewayTimeout, "Gateway timeout")
				return
			}
			httpx.Error(w, http.StatusBadGateway, "Bad gateway")
		},
	}, nil
}

// MaxIdleConnsPerUpstream is how many idle connections are kept per upstream.
const MaxIdleConnsPerUpstream = 256

// copyBufferSize is the size of the buffers response bodies are copied
// through, the same size the proxy would otherwise allocate per response.
const copyBufferSize = 32 * 1024

// bufferPool reuses copy buffers across responses and proxies, so copying a
// body does not allocate 32 KB for every request.
type bufferPool struct{ pool sync.Pool }

var buffers = &bufferPool{pool: sync.Pool{New: func() any {
	b := make([]byte, copyBufferSize)
	return &b
}}}

func (p *bufferPool) Get() []byte {
	return *p.pool.Get().(*[]byte)
}

func (p *bufferPool) Put(b []byte) {
	if cap(b) < copyBufferSize {
		return
	}
	b = b[:copyBufferSize]
	p.pool.Put(&b)
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
