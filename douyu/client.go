package douyu

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/songtianyi/rrframework/logs"
	"log"
	"net"
	"time"
)

type Client struct {
	conn            net.Conn
	HandlerRegister *HandlerRegister
	closed          chan struct{}

	rLock sync.Mutex
	wLock sync.Mutex
}

// Connect to douyu barrage server
func Connect(connStr string, handlerRegister *HandlerRegister) (*Client, error) {
	conn, err = net.Dial("tcp", connStr)
	if err != nil {
		return nil, err
	}

	// server connected
	client = &Client{
		conn: conn,
	}

	if handlerRegister == nil {
		client.HandlerRegister = CreateHandlerRegister()
	} else {
		client.HandlerRegister = handlerRegister
	}

	go c.heartbeat()
	return client, nil
}

// Send message to server
func (c *Client) Send(b []byte) (int, error) {
	return c.conn.Write(b)
}

// Receive message from server
func (c *Client) Receive() ([]byte, int, error) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	buf := make([]byte, 512)
	if _, err := io.ReadFull(c.conn, buf[:12]); err != nil {
		return buf, 0, err
	}

	// ignore 8 byte, first 4byte for message length
	pl := binary.LittleEndian.Uint32(buf[:4])

	// 2byte for message type
	code := binary.LittleEndian.Uint32(buf[8:10])

	if pl > 512 {
		// expand buffer
		buf = make([]byte, pl)
	}
	if _, err := io.ReadFull(c.conn, buf[:pl]); err != nil {
		return buf, code, err
	}
	return buf, code, nil
}

// Close connnection
func (c *Client) Close() error {
	c.closed <- struct{}{}
	return c.Conn.Close()
}

// JoinRoom for authentication
func (c *Client) JoinRoom(room int) error {
	loginMessage := NewMessage(nil, MESSAGE_TO_SERVER).
		SetField("type", "loginreq").
		SetField("roomid", rid)

	c.Send(loginMessage.Bytes())

	_, err := c.Receive()
	if err != nil {
		return err
	}

	joinMessage := NewMessage(nil, MESSAGE_TO_SERVER).
		SetField("type", "joingroup").
		SetField("rid", rid).
		SetField("gid", "-9999")

	b, err = c.Send(joinMessage.Bytes())
	if err != nil {
		return err
	}
	go c.serve()
	return nil
}

func (c *Client) serve() error {
loop:
	for {
		select {
		case <-c.closed:
			break loop
		default:
			b, code, err := c.Receive()
			if err != nil {
				logs.Error(err)
				break loop
			}

			// analize message
			msg := NewMessage(nil, MESSAGE_FROM_SERVER).Decode(b, code)
			handlers := c.HandlerRegister.Get(msg.GetStringField("type"))
			for _, v := range handlers {
				go v.Run(msg)
			}
		}
	}
}

func (c *Client) heartbeat() {
	tick := time.Tick(45 * time.Second)
	for {
		select {
		case <-tick:

			heartbeatMsg := NewMessage(nil, MESSGE_TO_SERVER).
				SetField("type", "keeplive").
				SetField("tick", timestamp)

			_, err := c.Send(heartbeatMsg.Bytes())
			if err != nil {
				log.Fatal("heartbeat failed " + err.Error())
			}
		}
	}
}
