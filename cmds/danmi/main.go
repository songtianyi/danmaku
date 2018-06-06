package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/songtianyi/danmaku/bilibili"
	"github.com/songtianyi/danmaku/douyu"
	"github.com/songtianyi/danmaku/panda"
	"github.com/songtianyi/rrframework/logs"
	"github.com/urfave/cli"
	"github.com/yanyiwu/gojieba"
)

var (
	jieba        = gojieba.NewJieba()
	string2Level = map[string]int{
		"EMERGENCY": logs.LevelEmergency,
		"ALERT":     logs.LevelAlert,
		"CRITICAL":  logs.LevelCritical,
		"ERROR":     logs.LevelError,
		"WARNING":   logs.LevelWarning,
		"NOTICE":    logs.LevelNotice,
		"INFO":      logs.LevelInformational,
		"DEBUG":     logs.LevelDebug,
	}
)

func chatmsg(msg *douyu.Message) {
	level := msg.GetStringField("level")
	nn := msg.GetStringField("nn")
	txt := msg.GetStringField("txt")

	// contents := jieba.CutAll(txt)

	// logs.Info(fmt.Sprintf("level(%s) - %s >>> %s\n%s", level, nn, txt, string(contents)))
	// logs.Info(fmt.Sprintf("level(%s) - %s >>> %s\n%s", level, nn, txt, strings.Join(contents, "/")))
	logs.Info(fmt.Sprintf("level(%s) - %s >>> %s", level, nn, txt))

}

func douyuHandler(ctx *cli.Context) error {
	server := ctx.String("s")
	client, err := douyu.Connect(server, nil)
	if err != nil {
		return err
	}

	client.HandlerRegister.Add("chatmsg", douyu.Handler(chatmsg), "chatmsg")
	if err := client.JoinRoom(ctx.Int("rid")); err != nil {
		logs.Error(fmt.Sprintf("Join room fail, %s", err.Error()))
		return err
	}
	client.Serve()
	return nil
}

func danmu(msg *bilibili.Message) {
	logs.Debug(">>>", string(msg.Bytes()))
}

func bilibiliHandler(ctx *cli.Context) error {
	prefix := ctx.String("s")
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	rid := ctx.String("rid")
	uid := ctx.Int("uid")
	client, err := bilibili.Connect(prefix+rid, uid, nil)
	if err != nil {
		return err
	}
	client.HandlerRegister.Add(bilibili.DANMU_MSG, bilibili.Handler(danmu), "danmu")
	client.Serve()
	return nil
}

func pandaTV(msg *panda.DecodedMessage) {
	logs.Debug("(%s) - %s >>> %s", msg.Type, msg.Nickname, msg.Content)
}

func pandaHandler(ctx *cli.Context) error {
	// uri, handlerRegister
	prefix := ctx.String("s")
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	rid := ctx.String("rid")
	client, err := panda.Connect(prefix+rid, nil)
	if err != nil {
		logs.Error(err)
		return err
	}
	client.HandlerRegister.Add(panda.DANMU_MSG, panda.Handler(pandaTV), "danmu")
	client.Serve()

	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "A cli tool to stat danmu messages."
	app.Version = "1.0.0"
	app.Commands = []cli.Command{
		{
			Name:   "douyu",
			Usage:  "connect to douyu danmu message server",
			Action: douyuHandler,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server, s",
					Value: "openbarrage.douyutv.com:8601",
					Usage: "douyu danmu message server address",
				},
				cli.IntFlag{
					Name:  "room, rid",
					Value: 667351,
					Usage: "douyu room id",
				},
			},
		},
		{
			Name:   "bilibili",
			Usage:  "connect to bilibili danmu message api",
			Action: bilibiliHandler,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server, s",
					Value: "https://live.bilibili.com/",
					Usage: "bilibili danmu api prefix",
				},
				cli.IntFlag{
					Name:  "room, rid",
					Value: 28645,
					Usage: "bilibili room id",
				},
				cli.IntFlag{
					Name:  "user, uid",
					Value: -1,
					Usage: "bilibili user id",
				},
			},
		},
		{
			Name:   "panda",
			Usage:  "connect to pandaTV danmu message api",
			Action: pandaHandler,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server, s",
					Value: "https://www.panda.tv/",
					Usage: "pandaTV danmu api prefix",
				},
				cli.IntFlag{
					Name:  "room, rid",
					Value: 66666,
					Usage: "pandaTV room id",
				},
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log, l",
			Value: "DEBUG",
			Usage: "log level settings, case insensitive, {" +
				"EMERGENCY|ALERT|CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG}",
		},
	}
	app.Action = func(ctx *cli.Context) error {
		if v, ok := string2Level[ctx.String("l")]; ok {
			logs.SetLevel(v)
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	return
}
