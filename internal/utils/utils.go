package utils

import "net/url"

func ConstructURL(baseURL, path string) (*url.URL, error) {
	return url.Parse(baseURL + path)
}
