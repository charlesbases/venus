package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/charlesbases/venus/plugins"
)

var app = &cli.App{
	Name:      filepath.Base(os.Args[0]),
	Usage:     "video website crawler",
	ArgsUsage: "[link]",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "f",
			Usage: "load video link from file",
			Value: "index.txt",
		},
		cli.UintFlag{
			Name:  "c",
			Usage: "number of videos downloaded at the same time",
			Value: 10,
		},
	},
	Action: func(ctx *cli.Context) error {
		cr := plugins.Crawler(ctx)
		defer cr.Stop()
		return cr.Start()
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
