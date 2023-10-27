package http2tcp

import (
	"context"
	"github.com/MikeLINGxZ/http2tcp/internal/utils"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"sync"
	"time"
)

type Connection struct {
	id        string
	wsConn    *websocket.Conn
	tcpConn   net.Conn
	isServer  bool
	cancelCtx *utils.CancelAllContext
	wg        sync.WaitGroup
	isClose   bool
}

func NewConnection(id string, wsConn *websocket.Conn, tcpConn net.Conn, isServer bool) *Connection {
	return &Connection{
		id:        id,
		wsConn:    wsConn,
		tcpConn:   tcpConn,
		isServer:  isServer,
		cancelCtx: utils.WithCancelAll(context.Background()),
		wg:        sync.WaitGroup{},
		isClose:   false,
	}
}

func (c *Connection) Proxy() {

	log.Printf("[connection] Proxy | start proxy id: %s \n", c.id)
	defer log.Printf("[connection] Proxy | done proxy id: %s \n", c.id)

	defer c.close()

	if !c.isServer {
		go c.keepPing()
	}
	go c.http2tcp()
	go c.tcp2http()

	// for wait wg.Add
	time.Sleep(time.Second * 5)
	c.wg.Wait()
}

func (c *Connection) close() {
	defer c.cancelCtx.CancelAll()

	if c.isClose {
		return
	}

	err := c.wsConn.Close()
	if err != nil {
		log.Printf("[connection] close | ws conn error: %s \n", err.Error())
	}
	err = c.tcpConn.Close()
	if err != nil {
		log.Printf("[connection] close | tcp conn error: %s \n", err.Error())
	}
	c.isClose = true
}

func (c *Connection) http2tcp() {
	c.wg.Add(1)
	defer c.wg.Done()
	defer c.cancelCtx.CancelAll()

	done, err := c.cancelCtx.GetDoneCh()
	if err != nil {
		log.Printf("[connection] http2tcp | get done ch error: %s \n", err.Error())
		return
	}

	for {
		select {
		case <-done:
			return
		default:
			msgType, msgBytes, err := c.wsConn.ReadMessage()
			if err != nil {
				log.Printf("[connection] http2tcp | read ws msg error: %s \n", err.Error())
				return
			}
			if msgType == websocket.PingMessage {
				err := c.wsConn.WriteMessage(websocket.PongMessage, []byte(""))
				if err != nil {
					log.Printf("[connection] http2tcp | write ws pong error: %s \n", err.Error())
					return
				}
				continue
			}
			if msgType != websocket.BinaryMessage {
				continue
			}
			_, err = c.tcpConn.Write(msgBytes)
			if err != nil {
				log.Printf("[connection] http2tcp | write tcp msg error: %s \n", err.Error())
				return
			}
		}
	}
}

func (c *Connection) tcp2http() {
	c.wg.Add(1)
	defer c.wg.Done()
	defer c.cancelCtx.CancelAll()

	done, err := c.cancelCtx.GetDoneCh()
	if err != nil {
		log.Printf("[connection] tcp2http | get done ch error: %s \n", err.Error())
		return
	}

	for {
		select {
		case <-done:
			return
		default:
			buf := make([]byte, 4096)
			n, err := c.tcpConn.Read(buf)
			if err != nil {
				log.Printf("[connection] tcp2http | read tcp msg error: %s \n", err.Error())
				return
			}
			buf = buf[:n]
			err = c.wsConn.WriteMessage(websocket.BinaryMessage, buf)
			if err != nil {
				log.Printf("[connection] tcp2http | write ws msg error: %s \n", err.Error())
				return
			}
		}
	}
}

func (c *Connection) keepPing() {
	c.wg.Add(1)
	defer c.wg.Done()
	defer c.cancelCtx.CancelAll()

	done, err := c.cancelCtx.GetDoneCh()
	if err != nil {
		log.Printf("[connection] keepPing | get done ch error: %s \n", err.Error())
		return
	}
	for {
		time.Sleep(5)
		select {
		case <-done:
			return
		default:
			err := c.wsConn.WriteMessage(websocket.PingMessage, []byte(""))
			if err != nil {
				log.Printf("[connection] keepPing | write ping msg error: %s \n", err.Error())
				return
			}
		}
	}
}
