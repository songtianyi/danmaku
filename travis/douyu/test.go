package main

import (
	"fmt"
	"github.com/songtianyi/barrage/douyu"
	"github.com/songtianyi/rrframework/logs"
	"github.com/yanyiwu/gojieba"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	jieba = gojieba.NewJieba()
)

func ltp(txt string) ([]byte, error) {
	km := url.Values{}
	km.Add("api_key", "E1v3e0N2o4yz6WdSneCAhY7JqZnYea4TDeUKjvgy")
	km.Add("text", txt)
	km.Add("pattern", "all")
	km.Add("format", "conll")

	uri := "http://api.ltp-cloud.com/analysis/?" + km.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func chatmsg(msg *douyu.Message) {
	level := msg.GetStringField("level")
	nn := msg.GetStringField("nn")
	txt := msg.GetStringField("txt")

	//contents := ltp(txt)
	contents := jieba.CutAll(txt)

	//logs.Info(fmt.Sprintf("level(%s) - %s >>> %s\n%s", level, nn, txt, string(contents)))
	logs.Info(fmt.Sprintf("level(%s) - %s >>> %s\n%s", level, nn, txt, strings.Join(contents, "/")))

}

func main() {
	defer jieba.Free()
	client, err := douyu.Connect("openbarrage.douyutv.com:8601", nil)
	if err != nil {
		logs.Error(err)
		return
	}

	client.HandlerRegister.Add("chatmsg", douyu.Handler(chatmsg), "chatmsg")
	if err := client.JoinRoom(667351); err != nil {
		//if err := client.JoinRoom(532152); err != nil {
		logs.Error(fmt.Sprintf("Join room fail, %s", err.Error()))
		return
	}
	client.Serve()
}
