package http2tcp

import "net/http"

func ProxyForHttpHandler(auth string) func(writer http.ResponseWriter, request *http.Request) {
	s := NewServer(&ServerConfig{
		Auth: auth,
	})
	return s.proxy
}
