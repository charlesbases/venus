package plugins

import (
	"os"
	"strings"

	"github.com/charlesbases/venus/sender"
	"github.com/charlesbases/venus/types"
	"github.com/charlesbases/venus/website"
	"github.com/charlesbases/venus/website/spank"
	"github.com/charlesbases/venus/website/xvideos"
)

// domain 域名正则
var domain = types.MustCompile(`https://([^/]+)`)

var hookFunc = map[string]website.WebHookFunc{
	"www.xvideos.com": xvideos.New,
	"spankbang.com":   spank.New,
}

// producer .
type producer interface {
	Produce() ([]website.WebHook, error)
	Enabled() bool
}

var _ producer = (*argsProducer)(nil)

type argsProducer []string

// Enabled .
func (a argsProducer) Enabled() bool {
	return len(a) != 0
}

// Produce .
func (a argsProducer) Produce() ([]website.WebHook, error) {
	var res = make([]website.WebHook, 0, len(a))
	for i := range a {
		if v := domain.FindSubString(a[i]); len(v) != 0 {
			if hook, found := hookFunc[v]; found {
				res = append(res, hook(types.Link(a[i])))
			}
		}
	}
	return res, nil
}

var _ producer = (*fileProducer)(nil)

type fileProducer string

// Enabled .
func (f fileProducer) Enabled() bool {
	return len(f) != 0
}

// Produce .
func (f fileProducer) Produce() ([]website.WebHook, error) {
	file, err := os.Open(string(f))
	if err != nil {
		return nil, err
	}

	var links argsProducer
	sender.ReadLine(
		func(v string) (isBreak bool) {
			if !strings.HasPrefix(strings.TrimSpace(v), "#") {
				links = append(links, v)
			}
			return false
		},
	)(file)
	file.Close()

	return links.Produce()
}
