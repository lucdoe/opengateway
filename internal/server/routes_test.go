// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/lucdoe/opengateway/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namedProxy answers every request with the name of the service it stands
// for, taken from its upstream URL, so a test sees which route matched.
func namedProxy(target string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(target)) })
}

func routesFor(t *testing.T, services map[string]config.Service) *Server {
	t.Helper()
	cfg := &config.Config{Services: services}
	s, err := NewServer(cfg, mux.NewRouter(), func(target string, _ time.Duration) (http.Handler, error) {
		return namedProxy(target), nil
	}, Middlewares{}, nil)
	require.NoError(t, err)
	return s
}

func endpoint(method, path string) config.Endpoint {
	return config.Endpoint{HTTPMethod: method, Path: path}
}

func TestTheMostSpecificRouteWinsWhateverTheConfigOrder(t *testing.T) {
	s := routesFor(t, map[string]config.Service{
		// Listed least specific first, on purpose.
		"catch-all": {URL: "http://catch-all", Endpoints: []config.Endpoint{endpoint("*", "/*")}},
		"files":     {URL: "http://files", Endpoints: []config.Endpoint{endpoint("*", "/files/*")}},
		"images":    {URL: "http://images", Endpoints: []config.Endpoint{endpoint("GET", "/files/images/*")}},
		"special":   {URL: "http://special", Endpoints: []config.Endpoint{endpoint("GET", "/files/special")}},
		"uploads":   {URL: "http://uploads", Endpoints: []config.Endpoint{endpoint("POST", "/files/*")}},
	})

	for _, tc := range []struct{ method, path, want string }{
		{"GET", "/files/special", "http://special"},       // exact before prefix
		{"GET", "/files/images/cat.png", "http://images"}, // longer prefix first
		{"GET", "/files/images", "http://images"},         // a prefix matches itself
		{"DELETE", "/files/images/cat.png", "http://files"},
		{"GET", "/files/report.pdf", "http://files"},
		{"POST", "/files/report.pdf", "http://uploads"}, // a named method before *
		{"GET", "/files", "http://files"},
		{"GET", "/elsewhere/deep", "http://catch-all"},
		{"PATCH", "/", "http://catch-all"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := serve(s, tc.method, tc.path)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tc.want, w.Body.String())
		})
	}
}

func TestACatchAllLeavesTheHealthCheckAlone(t *testing.T) {
	s := routesFor(t, map[string]config.Service{"all": {URL: "http://all", Endpoints: []config.Endpoint{endpoint("*", "/*")}}})

	w := serve(s, "GET", HealthPath)

	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestAPrefixRouteWithASubpathStaysInsideIt(t *testing.T) {
	s := routesFor(t, map[string]config.Service{"api": {URL: "http://api", Subpath: "/api", Endpoints: []config.Endpoint{endpoint("GET", "/*")}}})

	assert.Equal(t, http.StatusOK, serve(s, "GET", "/api").Code)
	assert.Equal(t, http.StatusOK, serve(s, "GET", "/api/people/7").Code)
	assert.Equal(t, http.StatusNotFound, serve(s, "GET", "/apiary").Code, "the prefix ends at a segment")
	assert.Equal(t, http.StatusNotFound, serve(s, "GET", "/other").Code)
}

func TestUpstreamPathRewritesWhatIsSentUpstream(t *testing.T) {
	var received string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.URL.RequestURI()
	}))
	t.Cleanup(upstream.Close)

	cfg := &config.Config{Services: map[string]config.Service{
		"accounts": {URL: upstream.URL + "/base", Subpath: "/api", Endpoints: []config.Endpoint{
			{HTTPMethod: "GET", Path: "/users/{id}", UpstreamPath: "/v2/accounts/{id}"},
			{HTTPMethod: "GET", Path: "/users/{id:[0-9]+}/files/*", UpstreamPath: "/storage/{id}/*"},
			{HTTPMethod: "GET", Path: "/legacy/*", UpstreamPath: "/new"},
			{HTTPMethod: "GET", Path: "/as-is/{id}"},
		}},
	}}
	s, err := NewServer(cfg, mux.NewRouter(), proxy.NewHandler, Middlewares{}, nil)
	require.NoError(t, err)

	for gateway, want := range map[string]string{
		"/api/users/7?fields=name":    "/base/v2/accounts/7?fields=name",
		"/api/users/7/files/a/b.txt":  "/base/storage/7/a/b.txt",
		"/api/users/7/files":          "/base/storage/7",
		"/api/legacy/anything/at/all": "/base/new",
		"/api/as-is/7":                "/base/api/as-is/7", // no UpstreamPath: the gateway path, as before
		"/api/users/caf%C3%A9":        "/base/v2/accounts/caf%C3%A9",
	} {
		t.Run(gateway, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.Router.ServeHTTP(w, httptest.NewRequest("GET", gateway, nil))
			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, want, received)
		})
	}
}

func TestRemainder(t *testing.T) {
	for _, tc := range []struct {
		path     string
		segments int
		want     string
	}{
		{"/api/files/a/b", 2, "/a/b"},
		{"/api/files", 2, ""},
		{"/api/files/", 2, "/"},
		{"/anything", 0, "/anything"},
		{"/api", 3, ""},
	} {
		assert.Equal(t, tc.want, remainder(tc.path, tc.segments), "%s after %d", tc.path, tc.segments)
	}
}
