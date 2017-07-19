package main

import (
	"github.com/songtianyi/barrage/huomao"
	"github.com/songtianyi/rrframework/logs"
)

func danmu(msg *huomao.DecodedMessage) {
	logs.Debug("(%s) - %s >>> %s", msg.Type, msg.Nickname, msg.Content)
}

func main() {
	// uri, handlerRegister
	client, err := huomao.Connect("https://www.huomao.com/8952", nil)
	if err != nil {
		logs.Error(err)
		return
	}
	client.HandlerRegister.Add(huomao.DANMU_MSG, huomao.Handler(danmu), "danmu")
	client.Serve()
}
