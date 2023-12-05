package services

import (
	"net/http/httputil"
	"net/url"
)

func NewProxy() *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "localhost:3000",
	})

	return proxy
}
