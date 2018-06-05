package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "A cli tool to "
	app.Version = "1.0.0"
	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "run workflow which described in yaml, if the ",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "mongo, mgo",
					Value: "mongodb://localhost:27017",
					Usage: "mongo db uri",
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	return
}
