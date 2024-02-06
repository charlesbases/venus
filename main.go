package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/charlesbases/venus/logger"
	"github.com/charlesbases/venus/regexp"
	"github.com/charlesbases/venus/website"
	"github.com/charlesbases/venus/website/spankbang"
	"github.com/charlesbases/venus/website/xvideos"
)

// 网站正则
var domain = regexp.New(`https://([^/]+)`)

// homeHook 视频网站首页
var homeHook = map[string]func(link website.Link) website.WebHook{
	"www.xvideos.com": xvideos.New,
	"spankbang.com":   spankbang.New,
}

var app = &cli.App{
	Name:      "viper",
	Usage:     "video download tool",
	ArgsUsage: "[link]",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "f",
			Usage: "load video link from file",
			Value: "index.txt",
		},
		cli.IntFlag{
			Name:  "c",
			Usage: "number of videos downloaded at the same time",
			Value: 10,
		},
	},
	Action: func(ctx *cli.Context) error {
		website.SetConcurrent(ctx.Int("c"))

		if err := logger.Create(); err != nil {
			return err
		}
		defer logger.Close()

		var links []string
		if len(ctx.Args()) != 0 {
			// load link with args
			links = os.Args
		} else {
			// load links with file
			file, err := os.Open(ctx.String("f"))
			if err != nil {
				return err
			}

			website.ReadLine(
				func(line string) (isBreak bool) {
					if !strings.HasPrefix(strings.TrimSpace(line), "#") {
						links = append(links, line)
					}
					return false
				},
			)(file)
			file.Close()
		}

		return website.Crawler(parseHookWithLink(links)...)
	},
}

// parseHookWithLink .
func parseHookWithLink(links []string) []website.WebHook {
	var hooks = make([]website.WebHook, 0, len(links))
	for _, link := range links {
		if webhome := domain.FindSubString(link); len(webhome) != 0 {
			if hook, found := homeHook[webhome]; found {
				hooks = append(hooks, hook(website.Link(link)))
			}
		}
	}
	return hooks
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
