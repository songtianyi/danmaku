package douyu

import (
	"encoding/binary"
	"fmt"
	"github.com/songtianyi/rrframework/logs"
	"io"
	"log"
	"net"
	"sync"
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
	conn, err := net.Dial("tcp", connStr)
	if err != nil {
		return nil, err
	}

	logs.Info(fmt.Sprintf("%s connected.", connStr))

	// server connected
	client := &Client{
		conn: conn,
	}

	if handlerRegister == nil {
		client.HandlerRegister = CreateHandlerRegister()
	} else {
		client.HandlerRegister = handlerRegister
	}

	go client.heartbeat()
	return client, nil
}

// Send message to server
func (c *Client) Send(b []byte) (int, error) {
	c.wLock.Lock()
	defer c.wLock.Unlock()
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

	// 12 bytes header
	// 4byte for packet length
	pl := binary.LittleEndian.Uint32(buf[:4])

	// ignore buf[4:8]

	// 2byte for message type
	code := binary.LittleEndian.Uint16(buf[8:10])

	// 1byte for secret
	// 1byte for reserved

	// body content length(include ENDING)
	cl := pl - 8

	if cl > 512 {
		// expand buffer
		buf = make([]byte, cl)
	}
	if _, err := io.ReadFull(c.conn, buf[:cl]); err != nil {
		return buf, int(code), err
	}
	// exclude ENDING
	return buf[:cl-1], int(code), nil
}

// Close connnection
func (c *Client) Close() error {
	c.closed <- struct{}{}
	return c.conn.Close()
}

// JoinRoom
func (c *Client) JoinRoom(room int) error {
	loginMessage := NewMessage(nil, MESSAGE_TO_SERVER).
		SetField("type", MSG_TYPE_LOGINREQ).
		SetField("roomid", room)

	logs.Info(fmt.Sprintf("joining room %d...", room))
	if _, err := c.Send(loginMessage.Encode()); err != nil {
		return err
	}

	b, code, err := c.Receive()
	if err != nil {
		return err
	}

	// TODO assert(code == MESSAGE_FROM_SERVER)
	logs.Info(fmt.Sprintf("room %d joined", room))
	loginRes := NewMessage(nil, MESSAGE_FROM_SERVER).Decode(b, code)
	logs.Info(fmt.Sprintf("room %d live status %s", room, loginRes.GetStringField("live_stat")))

	joinMessage := NewMessage(nil, MESSAGE_TO_SERVER).
		SetField("type", "joingroup").
		SetField("rid", room).
		SetField("gid", "-9999")

	logs.Info(fmt.Sprintf("joining group %d...", -9999))
	_, err = c.Send(joinMessage.Encode())
	if err != nil {
		return err
	}
	logs.Info(fmt.Sprintf("group %d joined", -9999))
	return nil
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
				break loop
			}

			// analize message
			msg := NewMessage(nil, MESSAGE_FROM_SERVER).Decode(b, code)
			err, handlers := c.HandlerRegister.Get(msg.GetStringField("type"))
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

func (c *Client) heartbeat() {
	tick := time.Tick(45 * time.Second)
	for {
		select {
		case <-tick:
			heartbeatMsg := NewMessage(nil, MESSAGE_TO_SERVER).
				SetField("type", "keeplive").
				SetField("tick", time.Now().Unix())

			_, err := c.Send(heartbeatMsg.Encode())
			if err != nil {
				log.Fatal("heartbeat failed, " + err.Error())
			}
		}
	}
}
