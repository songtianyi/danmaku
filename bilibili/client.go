package bilibili

import (
	"encoding/binary"
	"fmt"
	"github.com/songtianyi/rrframework/logs"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	conn            net.Conn
	HandlerRegister *HandlerRegister
	closed          chan struct{}
	roomid          int
	uid             int

	rLock sync.Mutex
	wLock sync.Mutex
}

func Connect(uri string, uid int, handlerRegister *HandlerRegister) (*Client, error) {

	roomStr, err := GetRoomId(uri)
	if err != nil {
		return nil, err
	}
	server, state, err := GetBarrageServerAndLiveState(roomStr)
	if err != nil {
		return nil, err
	}
	server += ":788"
	conn, err := net.DialTimeout("tcp", server, 10*time.Second)
	if err != nil {
		return nil, err
	}

	logs.Info(fmt.Sprintf("%s connected, live status %s", server, state))

	roomid, err := strconv.Atoi(roomStr)
	if err != nil {
		return nil, err
	}
	if uid < 0 {
		uid = RandUser()
	}

	client := &Client{
		conn:   conn,
		roomid: roomid,
		uid:    uid,
	}
	if handlerRegister == nil {
		client.HandlerRegister = CreateHandlerRegister()
	} else {
		client.HandlerRegister = handlerRegister
	}

	handshake := NewHandshakeMessage(roomid, uid)
	if _, err := client.Send(handshake.Encode()); err != nil {
		return nil, err
	}

	go client.heartbeat()
	return client, nil
}

func (c *Client) Send(b []byte) (int, error) {
	c.wLock.Lock()
	defer c.wLock.Unlock()
	return c.conn.Write(b)
}

func (c *Client) Receive() ([]byte, int, error) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	buf := make([]byte, 512)
	if _, err := io.ReadFull(c.conn, buf[:HEADER_LENGTH]); err != nil {
		return buf, -1, err
	}

	// header
	// 4byte for packet length
	pl := binary.BigEndian.Uint32(buf[:4])

	// ignore buf[4:6] and buf[6:8]
	code := int(binary.BigEndian.Uint32(buf[8:12]))
	// ignore buf[12:16]

	// body content length
	cl := pl - HEADER_LENGTH

	if cl > 512 {
		// expand buffer
		buf = make([]byte, cl)
	}
	if _, err := io.ReadFull(c.conn, buf[:cl]); err != nil {
		return buf, code, err
	}
	return buf[:cl], code, nil
}

// Close connnection
func (c *Client) Close() error {
	c.closed <- struct{}{} // receive
	c.closed <- struct{}{} // heartbeat
	return c.conn.Close()
}

func (c *Client) heartbeat() {
	tick := time.Tick(30 * time.Second)
loop:
	for {
		select {
		case <-c.closed:
			break loop
		case <-tick:
			handshake := NewHeartbeatMessage(c.roomid, c.uid)

			if _, err := c.conn.Write(handshake.Encode()); err != nil {
				logs.Error("heartbeat failed, " + err.Error())
			}
		}
	}
}

func (c *Client) Serve() {
loop:
	for {
		select {
		case <-c.closed:
			break loop
		default:
			b, code, err := c.Receive()
			if err != nil {
				logs.Error(err)
				continue
			}
			if code == 8 {
				logs.Info("handshake ok")
				continue
			}
			msg := NewMessage(b, code).Decode()
			err, handlers := c.HandlerRegister.Get(msg.GetCmd())
			if err != nil {
				logs.Warn(err)
				continue
			}
			for _, v := range handlers {
				go v.Run(msg)
			}

		}
	}
}
