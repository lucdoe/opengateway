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
	url, err := url.Parse(target)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.ServeHTTP(w, r)
	return nil
}
