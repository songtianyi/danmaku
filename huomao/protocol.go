package huomao

import (
	"bytes"
	"encoding/binary"
	"fmt"
	//"strings"
	"regexp"
)

const (
	DANMU_MSG = "1"
)

type DecodedMessage struct {
	Type     string
	Nickname string
	Content  string
}

type Message struct {
	head []byte
	body []byte

	Decoded []*DecodedMessage
}

func NewHandshakeStageOneMessage() []byte {

	data := fmt.Sprintf(`{"sys":{"version":"0.1.6b","pomelo_version":"0.7.x","type":"pomelo-flash-tcp"}}`)

	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, int32(len(data))) // write body length
	lenBuf := buffer.Bytes()
	lenBuf[0] = 0x01

	return append(lenBuf, []byte(data)...)

}

func NewHandshakeStageTwoMessage() []byte {
	return []byte{0x02, 0x00, 0x00, 0x00}
}

func NewHandshakeStageThreeMessage(roomId string) []byte {
	// head fixedstring data
	data := fmt.Sprintf(`{"channelId":roomId,"log":true,"userId":""}`)
	dataBuf := append([]byte{0x00, 0x01, 0x20}, []byte("gate.gateHandler.lookupConnector")...)
	dataBuf = append(dataBuf, []byte(data)...)

	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, int32(len(data))) // write body length
	lenBuf := buffer.Bytes()
	lenBuf[0] = 0x04

	// 0x04 int24(len(data)) dataBuf
	return append(lenBuf, dataBuf...)

}

func NewHeartbeatMessage() *Message {

	message := &Message{
		head: []byte{0x00, 0x06, 0x00, 0x00},
	}
	return message

}

func NewMessage(b []byte) *Message {
	return &Message{
		body: b,
	}
}

func (msg *Message) Encode() []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, msg.head) // write head
	if msg.body != nil && len(msg.body) > 0 {
		binary.Write(buffer, binary.BigEndian, int32(len(msg.body))) // write body length
		binary.Write(buffer, binary.BigEndian, msg.body)             // write body
	}
	return buffer.Bytes()
}

func (msg *Message) Decode() *Message {
	// TODO
	s := string(msg.body)
	//fmt.Println(s)

	// split by "ack:0"
	//raw := strings.Split(s, "ack:0")
	//for _, v := range raw {
	//	if n := strings.Index(v, "{"); n > -1 {
	//		js := v[n:]
	//		// unmarshal json
	//		fmt.Println(js)
	//	}
	//}

	nickNameReg := regexp.MustCompile("\"nickName\":\"([^\"]*)\"")
	nickNames := nickNameReg.FindAllStringSubmatch(s, -1)
	typeReg := regexp.MustCompile("\"type\":\"([^\"]*)\"")
	types := typeReg.FindAllStringSubmatch(s, -1)
	contentReg := regexp.MustCompile("\"content\":\"([^\"]*)\"")
	contents := contentReg.FindAllStringSubmatch(s, -1)

	msg.Decoded = make([]*DecodedMessage, 0)
	for i, v := range types {
		if v[1] != "1" {
			continue
		}
		msg.Decoded = append(msg.Decoded, &DecodedMessage{
			Type:     v[1],
			Nickname: nickNames[i][1],
			Content:  contents[i][1],
		})
	}
	return msg
}

func (msg *Message) Bytes() []byte {
	return msg.body
}
