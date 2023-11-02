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

const defaultHost = "0.0.0.0"
const defaultPort = "8989"
const defaultPath = "/"

type Server struct {
	config *ServerConfig
}

func NewServer(config *ServerConfig) *Server {
	if config.Host == "" {
		config.Host = defaultHost
	}
	if config.Port == "" {
		config.Port = defaultPort
	}
	if config.Path == "" {
		config.Path = defaultPath
	}
	return &Server{config: config}
}

func (s *Server) Run() error {
	log.Printf("[server] Run | server run on: %s ,proxy path: %s", s.config.Host+":"+s.config.Port, s.config.Path)
	http.HandleFunc(s.config.Path, s.proxy)
	return http.ListenAndServe(s.config.Host+":"+s.config.Port, nil)
}

func (s *Server) proxy(writer http.ResponseWriter, request *http.Request) {

	headerUpgrade := request.Header.Get("Upgrade")
	headerConnection := request.Header.Get("Connection")
	if headerUpgrade == "" || headerConnection == "" {
		writeError(writer, errors.New("header error: "+headerUpgrade+" "+headerConnection))
		return
	}

	wsConn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		writeError(writer, err)
		return
	}
	tcpConn, err := s.getTcpConn(wsConn)
	if err != nil {
		wsConn.Close()
		writeError(writer, err)
		return
	}

	connection := NewConnection(fmt.Sprintf("%d", rand.Int63()), wsConn, tcpConn, s.config.WaitTime, true)
	go connection.Proxy()
}

func (s *Server) getTcpConn(wsConn *websocket.Conn) (net.Conn, error) {
	retryTime := 3
	for i := 0; i < retryTime; i++ {
		msgType, msgBytes, err := wsConn.ReadMessage()
		if err != nil {
			return nil, err
		}
		if msgType != websocket.TextMessage {
			continue
		}
		var target Target
		err = json.Unmarshal(msgBytes, &target)
		if err != nil {
			return nil, err
		}
		if target.Auth != s.config.Auth {
			return nil, errors.New("auth not match")
		}
		tcpConn, err := net.Dial("tcp", target.RemoteHost+":"+target.RemotePort)
		if err != nil {
			return nil, err
		}
		return tcpConn, nil
	}
	return nil, errors.New("[server] can not read target info after 3 times retry")
}
