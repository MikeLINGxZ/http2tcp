package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MikeLINGxZ/http2tcp"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
)

type Client struct {
	config   *http2tcp.ClientConfig
	connList []*conn
}

type conn struct {
	listener net.Listener
	target   *http2tcp.Target
}

func NewClient(config *http2tcp.ClientConfig) *Client {
	return &Client{config: config, connList: []*conn{}}
}

func (c *Client) Run() error {
	for _, target := range c.config.Targets {
		target := target
		log.Printf("[client] Run | remote: %s ---------------- local: %s", target.RemoteHost+":"+target.RemotePort, target.LocalHost+":"+target.LocalPort)
		listen, err := net.Listen("tcp", target.LocalHost+":"+target.LocalPort)
		if err != nil {
			return err
		}
		c.connList = append(c.connList, &conn{
			listener: listen,
			target:   target,
		})
	}

	// wait all tcp listener ready
	for _, conn := range c.connList {
		conn := conn
		go c.listen(conn)
	}
	ch := make(chan struct{})
	<-ch
	return nil
}

func (c *Client) listen(conn *conn) {
	for {
		tcpConn, err := conn.listener.Accept()
		if err != nil {
			log.Printf("[client] listen | accept tcp conn error: %s \n", err.Error())
			panic(err)
		}
		go c.proxy(conn.target, tcpConn)
	}
}

func (c *Client) proxy(target *http2tcp.Target, tcpConn net.Conn) {
	wsConn, err := c.getWsConn(target)
	if err != nil {
		tcpConn.Close()
		log.Printf("[client] proxy | get ws conn error: %s \n", err.Error())
		return
	}
	connection := http2tcp.NewConnection(fmt.Sprintf("%d", rand.Int63()), wsConn, tcpConn, false)
	go connection.Proxy()
}

func (c *Client) getWsConn(target *http2tcp.Target) (*websocket.Conn, error) {
	target.Auth = c.config.Auth
	wsConn, httpResponse, err := websocket.DefaultDialer.Dial(c.config.WebsocketServer, http.Header{})
	if err != nil {
		return nil, err
	}
	if httpResponse.StatusCode != 101 {
		bytes, err := io.ReadAll(httpResponse.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(bytes))
	}
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
