package main

import (
	"github.com/songtianyi/barrage/panda"
	"github.com/songtianyi/rrframework/logs"
)

func danmu(msg *panda.DecodedMessage) {
	logs.Debug("(%s) - %s >>> %s", msg.Type, msg.Nickname, msg.Content)
}

func main() {
	// uri, handlerRegister
	client, err := panda.Connect("https://www.panda.tv/66666", nil)
	if err != nil {
		logs.Error(err)
		return
	}
	client.HandlerRegister.Add(panda.DANMU_MSG, panda.Handler(danmu), "danmu")
	client.Serve()
}
