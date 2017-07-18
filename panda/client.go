package panda

import (
	//"encoding/binary"
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
	chatInfo        *ChatInfo

	rLock sync.Mutex
	wLock sync.Mutex
}

func Connect(uri string, handlerRegister *HandlerRegister) (*Client, error) {

	roomStr, err := GetRoomId(uri)
	if err != nil {
		return nil, err
	}

	state, err := GetBarrageLiveState(roomStr)
	if err != nil {
		return nil, err
	}

	chatInfo, err := GetBarrageChatInfo(roomStr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", chatInfo.Data.ChatAddrList[0], 10*time.Second)
	if err != nil {
		return nil, err
	}

	logs.Info(fmt.Sprintf("%s connected, live status %s", chatInfo.Data.ChatAddrList[0], state))

	roomid, err := strconv.Atoi(roomStr)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:     conn,
		roomid:   roomid,
		chatInfo: chatInfo,
	}

	if handlerRegister == nil {
		client.HandlerRegister = CreateHandlerRegister()
	} else {
		client.HandlerRegister = handlerRegister
	}

	handshake := NewHandshakeMessage(chatInfo)
	if _, err := client.Send(handshake.Encode()); err != nil {
		return nil, err
	}

	buf := make([]byte, 28)
	if _, err := io.ReadFull(client.conn, buf); err != nil {
		return nil, err
	}
	logs.Info("handshake ok")

	go client.heartbeat()
	return client, nil
}

func (c *Client) Send(b []byte) (int, error) {
	c.wLock.Lock()
	defer c.wLock.Unlock()
	return c.conn.Write(b)
}

func (c *Client) Receive() ([]byte, error) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	buf := make([]byte, 4096) // big buffer

	n, err := c.conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			return buf[:n], err
		}
	}
	return buf[:n], nil
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
			heartbeat := NewHeartbeatMessage()

			if _, err := c.conn.Write(heartbeat.Encode()); err != nil {
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
			b, err := c.Receive()
			if err != nil {
				logs.Error(err)
				continue
			}
			for _, dm := range NewMessage(b).Decode().Decoded {
				err, handlers := c.HandlerRegister.Get(dm.Type)
				if err != nil {
					logs.Warn(err)
					continue
				}
				for _, v := range handlers {
					go v.Run(dm)
				}
			}

		}
	}
}
