package spankbang

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charlesbases/venus/website"
)

var _ website.WebHook = (*spank)(nil)

// spank .
type spank struct {
	src  website.Link
	user *website.UserInfor
}

// UserInfor 用户视频信息
func (s *spank) UserInfor() (*website.UserInfor, error) {
	// spankbang.com 只支持视频下载
	return s.parseUserInfoFromVideoLink()
}

// ResponseStreamData .
type ResponseStreamData map[string]interface{}

// best 最优分辨率下载链接
func (r ResponseStreamData) best() string {
	for _, s := range []string{"4k", "1080p"} {
		if v, found := r[s]; found {
			if val, ok := v.([]interface{}); ok && len(val) > 0 {
				return fmt.Sprintf(`%v`, val[0])
			}
		}
	}
	return ""
}

// videoID .
func (r ResponseStreamData) videoID() string {
	if v, found := r["stream_raw_id"]; found {
		if id, ok := v.(float64); ok {
			return fmt.Sprintf(`%.f`, id)
		}
	}
	return ""
}

// parseUserInfoFromVideoLink .
func (s *spank) parseUserInfoFromVideoLink() (infor *website.UserInfor, err error) {
	var resp ResponseStreamData = make(map[string]interface{})
	if err := s.src.Fetch(
		website.ReadLine(
			func(line string) (isBreak bool) {
				if data := compileParseStreamDataFromVideoLink.FindSubString(line); len(data) != 0 {
					data = strings.ReplaceAll(data, `'`, `"`)
					err = json.Unmarshal([]byte(data), &resp)
					isBreak = true
				}
				return isBreak
			},
		),
		func(meta *website.Metadata) {
			meta.Command = website.NewCommand(
				exec.Command(
					"curl",
					"-H",
					"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
					s.src.String(),
				),
			)
			// meta.Header = map[string]string{
			// 	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
			// }
		},
	); err != nil {
		return nil, err
	}

	if hls := resp.best(); len(hls) != 0 {
		s.user.Videos = append(
			s.user.Videos, &website.Header{
				ID:   resp.videoID(),
				Link: website.Link(hls),
			},
		)
	}

	return s.user, nil
}

// ParseVideo 根据视频网页链接，获取下载地址
func (s *spank) ParseVideo(h *website.Header) (*website.Video, error) {
	return &website.Video{
		Header: h,
		HLink:  h.Link,
		Parts:  []website.Link{h.Link},
	}, nil
}

// New .
func New(src website.Link) website.WebHook {
	return &spank{
		src: src,
		user: &website.UserInfor{
			Root: "spankbang.com",
		},
	}
}
