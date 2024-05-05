package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func ReverseProxyHandler(originServerURL *url.URL) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		req.Host = originServerURL.Host
		req.URL.Host = originServerURL.Host
		req.URL.Scheme = originServerURL.Scheme
		req.RequestURI = ""

		originServerResponse, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(rw, "Failed to forward request: ", err)
			return
		}
		defer originServerResponse.Body.Close()

		for key, values := range originServerResponse.Header {
			for _, value := range values {
				rw.Header().Add(key, value)
			}
		}

		rw.WriteHeader(originServerResponse.StatusCode)

		io.Copy(rw, originServerResponse.Body)
	})
}
