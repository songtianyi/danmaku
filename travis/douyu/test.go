package main

import (
	"fmt"
	"github.com/songtianyi/barrage/douyu"
	"github.com/songtianyi/rrframework/logs"
)

func main() {
	client, err := douyu.Connect("openbarrage.douyutv.com:8601", nil)
	if err != nil {
		logs.Error(err)
		return
	}
	if err := client.JoinRoom(532152); err != nil {
		logs.Error(fmt.Sprintf("Join room fail, %s", err.Error()))
		return
	}
}
