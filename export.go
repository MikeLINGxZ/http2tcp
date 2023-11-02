package http2tcp

import "net/http"

func ProxyForHttpHandler(auth string, waitTime int) func(writer http.ResponseWriter, request *http.Request) {
	s := NewServer(&ServerConfig{
		Auth:     auth,
		WaitTime: waitTime,
	})
	return s.proxy
}
