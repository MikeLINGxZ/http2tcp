package http2tcp

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func writeError(writer http.ResponseWriter, err error) {
	_, err = writer.Write([]byte(err.Error()))
	if err != nil {
		log.Println("write error err:", err)
	}
}

type CancelAllContext struct {
	context.Context
	doneCh []chan struct{}
	lock   sync.RWMutex
	isDone bool
}

func WithCancelAll(ctx context.Context) *CancelAllContext {
	return &CancelAllContext{
		Context: ctx,
		doneCh:  []chan struct{}{},
		lock:    sync.RWMutex{},
		isDone:  false,
	}
}

func (c *CancelAllContext) GetDoneCh() (chan struct{}, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.isDone {
		return nil, errors.New("all context has been canceled")
	}

	ch := make(chan struct{}, 1)
	c.doneCh = append(c.doneCh, ch)
	return ch, nil
}

func (c *CancelAllContext) CancelAll() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.isDone {
		return
	}
	for _, ch := range c.doneCh {
		ch <- struct{}{}
		close(ch)
	}
	c.isDone = true
}
