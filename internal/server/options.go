package server

type option struct {
	auth    *string
	host    *string
	port    *string
	path    *string
	maxConn *int
}

func WithAuth(auth string) func(o *option) {
	return func(o *option) {
		o.auth = &auth
	}
}

func WithListenHost(host string) func(o *option) {
	return func(o *option) {
		o.host = &host
	}
}

func WithListenPort(port string) func(o *option) {
	return func(o *option) {
		o.port = &port
	}
}

func WithListenPath(path string) func(o *option) {
	return func(o *option) {
		o.path = &path
	}
}

func WithMaxConn(max int) func(o *option) {
	return func(o *option) {
		o.maxConn = &max
	}
}
