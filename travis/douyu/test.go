package main

import (
	"fmt"
	"github.com/songtianyi/barrage/douyu"
	"github.com/songtianyi/rrframework/logs"
	"io/ioutil"
	"net/http"
	"net/url"
)

func chatmsg(msg *douyu.Message) {
	level := msg.GetStringField("level")
	nn := msg.GetStringField("nn")
	txt := msg.GetStringField("txt")

	km := url.Values{}
	km.Add("api_key", "E1v3e0N2o4yz6WdSneCAhY7JqZnYea4TDeUKjvgy")
	km.Add("text", txt)
	km.Add("pattern", "all")
	km.Add("format", "conll")

	uri := "http://api.ltp-cloud.com/analysis/?" + km.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		logs.Info(fmt.Sprintf("level(%s) - %s >>> %s | %s", level, nn, txt, err))
		return
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Info(fmt.Sprintf("level(%s) - %s >>> %s | %s", level, nn, txt, err))
		return
	}
	logs.Info(fmt.Sprintf("level(%s) - %s >>> %s\n%s", level, nn, txt, string(contents)))

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
