package http2tcp

func ProxyForHttpHandler() {

}

func ProxyForGinHandler() {

}

type ServerConfig struct {
	Host string
	Port string
	Path string
	Auth string
}

type ClientConfig struct {
	WebsocketServer string
	Auth            string
	Targets         []*Target
}

type Target struct {
	LocalHost  string
	LocalPort  string
	RemoteHost string `json:"remote_host"`
	RemotePort string `json:"remote_port"`
	Auth       string `json:"auth"`
}
