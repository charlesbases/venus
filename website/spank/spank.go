package spank

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charlesbases/venus/sender"
	"github.com/charlesbases/venus/types"
	"github.com/charlesbases/venus/website"
)

const domain = "spankbang.com"

// 视频分辨率优先级
var tendency = []string{"4k", "1080p"}

var _ website.WebHook = (*spank)(nil)

// spank spankbang.com 只支持单视频下载
type spank struct {
	base   types.LinkPath
	sender sender.Sender
	meta   []sender.Option
}

// New .
func New(link types.LinkPath) website.WebHook {
	return &spank{
		base:   link,
		sender: sender.CommandSender,
		meta: []sender.Option{
			sender.Header(
				map[string]string{
					"User-Agent": sender.UserAgentGoogleChrome,
				},
			),
		},
	}
}

// Sender .
func (s *spank) Sender() sender.Sender {
	return s.sender
}

// UserInfor .
func (s *spank) UserInfor() (*website.UserInfor, error) {
	return s.fromVideoLink()
}

// StreamData .
type StreamData map[string]interface{}

// get .
func (d StreamData) get(key string) string {
	if val, found := d[key]; found {
		switch val.(type) {
		case string:
			return val.(string)
		case float64:
			return fmt.Sprintf(`%.f`, val)
		case []interface{}:
			v := val.([]interface{})
			if len(v) > 0 {
				return fmt.Sprintf(`%v`, v[0])
			}
		}
	}

	return ""
}

// best .
func (d StreamData) best() string {
	for i := range tendency {
		if v := d.get(tendency[i]); len(v) != 0 {
			return v
		}
	}
	return ""
}

// fromVideoLink .
func (s *spank) fromVideoLink() (user *website.UserInfor, err error) {
	var m3u8 StreamData = make(map[string]interface{})

	if err := s.Sender().SendRequest(
		s.base.String(), sender.ReadLine(
			func(line string) (isBreak bool) {
				if data := cpStreamData.FinsSubStringBeforePrefix(line, "var stream_data"); len(data) != 0 {
					data = strings.ReplaceAll(data, `'`, `"`)
					err = json.Unmarshal([]byte(data), &m3u8)
					isBreak = true
				}
				return isBreak
			},
		),
		s.meta...,
	); err != nil {
		return nil, err
	}

	user = &website.UserInfor{Root: domain}
	if link := m3u8.best(); len(link) != 0 {
		user.Videos = append(
			user.Videos, &website.Header{
				VideoID:  m3u8.get("stream_raw_id"),
				LinkPath: types.Link(link),
			},
		)
	}
	return user, nil
}

// ParseHeader .
func (s *spank) ParseHeader(h *website.Header) (*website.Video, error) {
	return &website.Video{
		Header:   h,
		Metadata: s.meta,
		Parts:    []types.LinkPath{h.LinkPath},
	}, nil
}
