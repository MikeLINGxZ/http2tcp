package http2tcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net"
	"net/http"
)

type Client struct {
	config  *ClientConfig
	proxies []*proxy
}

type proxy struct {
	listener net.Listener
	target   *Target
}

func NewClient(config *ClientConfig) *Client {
	return &Client{config: config, proxies: []*proxy{}}
}

func (c *Client) Run() error {
	for _, target := range c.config.Targets {
		target := target
		listener, err := net.Listen("tcp", target.LocalHost+":"+target.LocalPort)
		if err != nil {
			return err
		}
		c.proxies = append(c.proxies, &proxy{
			listener: listener,
			target:   target,
		})
	}
	for _, p := range c.proxies {
		p := p
		log.Printf("[client] remote: %s ------- local: %s \n", p.target.RemoteHost+":"+p.target.RemotePort, p.target.LocalHost+":"+p.target.LocalPort)
		go c.handlerProxy(p)
	}
	ch := make(chan struct{})
	<-ch
	return nil
}

func (c *Client) handlerProxy(proxy *proxy) {
	for {
		tcpConn, err := proxy.listener.Accept()
		if err != nil {
			log.Printf("[client] handlerProxy | accept tcp conn error: %s \n", err.Error())
			continue
		}
		wsConn, err := c.getWsConn(proxy.target)
		if err != nil {
			tcpConn.Close()
			log.Printf("[client] handlerProxy | get ws conn error: %s \n", err.Error())
			continue
		}
		connection := NewConnection(fmt.Sprintf("%d", rand.Int63()), wsConn, tcpConn, false)
		go connection.Proxy()
	}
}

func (c *Client) getWsConn(target *Target) (*websocket.Conn, error) {
	wsConn, response, err := websocket.DefaultDialer.Dial(c.config.WebsocketServer, http.Header{})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 101 {
		return nil, errors.New("code != 101")
	}
	target.Auth = c.config.Auth
	bytes, err := json.Marshal(target)
	if err != nil {
		return nil, err
	}
	err = wsConn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		return nil, err
	}
	return wsConn, nil
}
