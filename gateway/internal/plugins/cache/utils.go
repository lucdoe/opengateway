package cache

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func generateCacheKey(r *http.Request) string {
	serviceIdentifier := r.Context().Value("service") // Ensure these are set somewhere earlier in the middleware chain
	endpointIdentifier := r.Context().Value("endpoint")

	sortedParams := sortQueryParams(r.URL.Query())
	return fmt.Sprintf("%s:%s:%s:%s:%s", serviceIdentifier, endpointIdentifier, r.Method, r.URL.Path, sortedParams)
}

func sortQueryParams(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sortedParams []string
	for _, k := range keys {
		sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", k, params.Get(k)))
	}
	return strings.Join(sortedParams, "&")
}
