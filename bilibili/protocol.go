package bilibili

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/songtianyi/rrframework/config"
	"github.com/songtianyi/rrframework/logs"
)

const (
	HEADER_LENGTH = 16 // in bytes
	DEVICE_TYPE   = 1
	DEVICE        = 1
)

const (
	// cmd types
	DANMU_MSG = "DANMU_MSG"

	//
	SERVER_PORT = "2243"
)

type Message struct {
	body     []byte
	bodyType int32
}

func NewHandshakeMessage(roomid, uid int) *Message {

	data := fmt.Sprintf(`{"roomid":%d,"uid":%d}`, roomid, uid)
	message := &Message{
		body:     []byte(data),
		bodyType: 7,
	}
	return message

}

func NewHeartbeatMessage(room, uid int) *Message {

	data := fmt.Sprintf(`{"roomid":%d,"uid":%d}`, room, uid)
	message := &Message{
		body:     []byte(data),
		bodyType: 2,
	}
	return message

}

func NewMessage(b []byte, btype int) *Message {
	return &Message{
		body:     b,
		bodyType: int32(btype),
	}

}

func (msg *Message) Encode() []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, int32(len(msg.body)+HEADER_LENGTH)) // write package length
	binary.Write(buffer, binary.BigEndian, int16(HEADER_LENGTH))               // header length
	binary.Write(buffer, binary.BigEndian, int16(DEVICE_TYPE))
	binary.Write(buffer, binary.BigEndian, int32(msg.bodyType))
	binary.Write(buffer, binary.BigEndian, int32(DEVICE))
	binary.Write(buffer, binary.BigEndian, msg.body)
	return buffer.Bytes()
}

func (msg *Message) Decode() *Message {
	// TODO
	return msg
}

func (msg *Message) GetCmd() string {
	jc, err := rrconfig.LoadJsonConfigFromBytes(msg.body)
	if err != nil {
		logs.Error(err)
		return "INVALID"
	}
	cmd, err := jc.GetString("cmd")
	if err != nil {
		logs.Error(err)
		return "ERROR"
	}
	return cmd

}

func (msg *Message) Bytes() []byte {
	return msg.body
}
