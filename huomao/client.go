package huomao

import (
	//"encoding/binary"
	"fmt"
	"github.com/songtianyi/rrframework/logs"
	"io"
	"net"
	//"strconv"
	"sync"
	"time"
)

type Client struct {
	conn            net.Conn
	HandlerRegister *HandlerRegister
	closed          chan struct{}
	roomid          string
	chatInfo        *ChatInfo

	rLock sync.Mutex
	wLock sync.Mutex
}

func Connect(uri string, handlerRegister *HandlerRegister) (*Client, error) {

	state, roomId, err := GetBarrageLiveStateRoomId(uri)
	if err != nil {
		return nil, err
	}
	fmt.Println(state, roomId)

	chatInfo, err := GetBarrageChatInfo(roomId)
	if err != nil {
		return nil, err
	}
	fmt.Println(chatInfo)

	connStr := chatInfo.Data.Host + ":" + chatInfo.Data.Port
	conn, err := net.DialTimeout("tcp", connStr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	logs.Info(fmt.Sprintf("%s connected, live status %s", connStr, state))

	client := &Client{
		conn:     conn,
		roomid:   roomId,
		chatInfo: chatInfo,
	}

	if handlerRegister == nil {
		client.HandlerRegister = CreateHandlerRegister()
	} else {
		client.HandlerRegister = handlerRegister
	}

	stageOne := NewHandshakeStageOneMessage()
	if _, err := client.Send(stageOne); err != nil {
		return client, err
	}

	resOne, err := client.Receive()
	if err != nil {
		return client, err
	}
	logs.Info("stage one", string(resOne))

	stageTwo := NewHandshakeStageTwoMessage()
	if _, err := client.Send(stageTwo); err != nil {
		return client, err
	}

	stageThree := NewHandshakeStageThreeMessage(roomId)
	if _, err := client.Send(stageThree); err != nil {
		return client, err
	}
	resThree, err := client.Receive()
	if err != nil {
		return client, err
	}
	logs.Info("stage three", string(resThree))

	//go client.heartbeat()
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
