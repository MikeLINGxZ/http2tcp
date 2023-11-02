package http2tcp

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"sync"
	"time"
)

type websocketMsgChan struct {
	msgType int
	msg     []byte
}

type Connection struct {
	id         string
	wsConn     *websocket.Conn
	tcpConn    net.Conn
	isServer   bool
	cancelCtx  *CancelAllContext
	wg         sync.WaitGroup
	isClose    bool
	lastActive time.Time
	waitTime   time.Duration
	wsWriterCh chan *websocketMsgChan
}

func NewConnection(id string, wsConn *websocket.Conn, tcpConn net.Conn, waitTime int, isServer bool) *Connection {
	return &Connection{
		id:         id,
		wsConn:     wsConn,
		tcpConn:    tcpConn,
		isServer:   isServer,
		cancelCtx:  WithCancelAll(context.Background()),
		wg:         sync.WaitGroup{},
		isClose:    false,
		waitTime:   time.Duration(waitTime) * time.Second,
		wsWriterCh: make(chan *websocketMsgChan),
	}
}

func (c *Connection) Proxy() {

	log.Printf("[connection] Proxy | start proxy id: %s \n", c.id)
	defer log.Printf("[connection] Proxy | done proxy id: %s \n", c.id)

	defer c.close()

	if !c.isServer {
		go c.keepPing()
	}
	if c.waitTime > time.Second*0 {
		go c.checkActive()
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

	close(c.wsWriterCh)
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
	defer c.close()

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
				c.writeWebsocketMsg(websocket.PongMessage, []byte(""))
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
			c.lastActive = time.Now()
		}
	}
}

func (c *Connection) tcp2http() {
	c.wg.Add(1)
	defer c.wg.Done()
	defer c.close()

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
			c.writeWebsocketMsg(websocket.BinaryMessage, buf)
			c.lastActive = time.Now()
		}
	}
}

func (c *Connection) keepPing() {
	c.wg.Add(1)
	defer c.wg.Done()
	defer c.close()

	done, err := c.cancelCtx.GetDoneCh()
	if err != nil {
		log.Printf("[connection] keepPing | get done ch error: %s \n", err.Error())
		return
	}
	for {
		time.Sleep(time.Second * 5)
		select {
		case <-done:
			return
		default:
			c.writeWebsocketMsg(websocket.PingMessage, []byte(""))
		}
	}
}

func (c *Connection) keepWsWrite() {
	c.wg.Add(1)
	defer c.wg.Done()
	defer c.close()

	done, err := c.cancelCtx.GetDoneCh()
	if err != nil {
		log.Printf("[connection] keepWsWrite | get done ch error: %s \n", err.Error())
		return
	}
	for {
		select {
		case <-done:
			return
		case msg := <-c.wsWriterCh:
			err := c.wsConn.WriteMessage(msg.msgType, msg.msg)
			if err != nil {
				log.Printf("[connection] keepWsWrite | write ws msg error: %s \n", err.Error())
				return
			}
		}
	}
}

func (c *Connection) checkActive() {
	c.wg.Add(1)
	defer c.wg.Done()
	defer c.close()

	done, err := c.cancelCtx.GetDoneCh()
	if err != nil {
		log.Printf("[connection] keepPing | get done ch error: %s \n", err.Error())
		return
	}

	for {
		time.Sleep(time.Second * 10)
		select {
		case <-done:
			return
		default:
			duration := time.Now().Sub(c.lastActive)
			if duration > c.waitTime {
				log.Printf("[connection] checkActive | idle for too long, end the connection \n")
				return
			}
		}
	}
}

func (c *Connection) writeWebsocketMsg(msgType int, msg []byte) {
	c.wsWriterCh <- &websocketMsgChan{
		msgType: msgType,
		msg:     msg,
	}
}
