package cache

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func generateCacheKey(r *http.Request) string {
	path := r.URL.Path
	address := r.RemoteAddr
	return fmt.Sprintf("%s:%s:%s:%s", r.Method, path, sortQueryParams(r.URL.Query()), address)
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
