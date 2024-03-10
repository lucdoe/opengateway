package utils

import "net/url"

type URLConstructorI interface {
	ConstructURL(baseURL, path string) (*url.URL, error)
}

type URLConstructor struct{}

func (u *URLConstructor) ConstructURL(baseURL, path string) (*url.URL, error) {
	return url.Parse(baseURL + path)
}

func GatewayURLConstructor() URLConstructorI {
	return &URLConstructor{}
}
