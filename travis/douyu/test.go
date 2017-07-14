package main

import (
	//"strings"
	"fmt"
	"github.com/songtianyi/barrage/douyu"
	"github.com/songtianyi/rrframework/logs"
	//"github.com/yanyiwu/gojieba"
)

func chatmsg(msg *douyu.Message) {
	level := msg.GetStringField("level")
	nn := msg.GetStringField("nn")
	txt := msg.GetStringField("txt")

	//jieba := gojieba.NewJieba()
	//defer jieba.Free()
	//words := jieba.Cut(txt, true)
	//logs.Info(fmt.Sprintf("level(%s) - %s >>> %s | %s", level, nn, txt, strings.Join(words, "/")))
	logs.Info(fmt.Sprintf("level(%s) - %s >>> %s", level, nn, txt))
}

func main() {
	client, err := douyu.Connect("openbarrage.douyutv.com:8601", nil)
	if err != nil {
		logs.Error(err)
		return
	}

	client.HandlerRegister.Add("chatmsg", douyu.Handler(chatmsg), "chatmsg")
	if err := client.JoinRoom(288016); err != nil {
		//if err := client.JoinRoom(532152); err != nil {
		logs.Error(fmt.Sprintf("Join room fail, %s", err.Error()))
		return
	}
	client.Serve()
}
