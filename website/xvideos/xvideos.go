package xvideos

import (
	"strconv"
	"strings"

	"github.com/charlesbases/venus/sender"
	"github.com/charlesbases/venus/types"
	"github.com/charlesbases/venus/website"
)

const domain = "xvideos.com"

var webhome = types.Link("https://www.xvideos.com")

// 视频分辨率优先级
var tendency = []string{"hls-1080p"}

var _ website.WebHook = (*xvideos)(nil)

// xvideos .
type xvideos struct {
	base   types.LinkPath
	sender sender.Sender
}

// New .
func New(link types.LinkPath) website.WebHook {
	return &xvideos{
		base:   link,
		sender: sender.HTTPSender,
	}
}

// Sender .
func (x *xvideos) Sender() sender.Sender {
	return x.sender
}

// UserInfor .
func (x *xvideos) UserInfor() (*website.UserInfor, error) {
	// user
	if cpIsUserLink.MatchString(x.base.String()) {
		return x.fromUserLink()
	}

	// video
	if cpIsVideoLink.MatchString(x.base.String()) {
		return x.fromVideoLink()
	}

	// unknown
	return nil, types.NewSenderError(x.base.String(), website.ErrLinkType)
}

// UserVideosResponse .
type UserVideosResponse struct {
	Videos []*struct {
		ID int    `json:"id"`
		TF string `json:"tf"`
		U  string `json:"u"`
	} `json:"videos"`
}

// fromUserLink .
func (x *xvideos) fromUserLink() (*website.UserInfor, error) {
	user := &website.UserInfor{ID: cpParseUserIDFromLink.FindSubString(x.base.String()), Root: domain}

	// 获取视频列表
	var page int
	for {
		var resp = new(UserVideosResponse)

		// 格式错误
		if err := x.Sender().SendRequest(
			x.base.PathJoin("videos", "new", strconv.Itoa(page)).String(), sender.Unmarshal(resp),
		); err != nil {
			return nil, err
		}

		// 页码错误
		if len(resp.Videos) == 0 {
			break
		}

		for _, video := range resp.Videos {
			if suffix := cpParseVideoIDFromUserHomePage.FindSubString(video.U); len(suffix) != 0 {
				user.Videos = append(
					user.Videos, &website.Header{
						Title:    video.TF,
						VideoID:  strconv.Itoa(video.ID),
						LinkPath: webhome.PathJoin("video." + suffix),
					},
				)
			}
		}
		page++
	}

	return user, nil
}

// fromVideoLink .
func (x *xvideos) fromVideoLink() (*website.UserInfor, error) {
	return &website.UserInfor{
		Root: domain,
		Videos: []*website.Header{
			{
				VideoID:  cpParseVideoIDFromLink.FindSubString(x.base.String()),
				LinkPath: x.base,
			},
		},
	}, nil
}

// ParseHeader .
func (x *xvideos) ParseHeader(h *website.Header) (*website.Video, error) {
	video := &website.Video{Header: h}

	// parse hls link
	var hls types.LinkPath
	if err := x.Sender().SendRequest(
		h.LinkPath.String(), sender.ReadLine(
			func(v string) (isBreak bool) {
				if link := cpParseHLinkFromUserHomePage.FinsSubStringBeforePrefix(
					v, "html5player.setVideoHLS",
				); len(link) != 0 {
					hls = types.Link(link)
					return true
				}
				return false
			},
		),
	); err != nil {
		return nil, err
	}

	// 解析分辨率
	r := newResolution()
	x.Sender().SendRequest(
		hls.PathJoin("hls.m3u8").String(), sender.ReadLine(
			func(v string) (isBreak bool) {
				r.add(v)
				return false
			},
		),
	)

	// 视频分片下载列表
	if best := r.best(); len(best) != 0 {
		x.Sender().SendRequest(
			hls.PathJoin(best).String(), sender.ReadLine(
				func(v string) (isBreak bool) {
					if !strings.HasPrefix(v, "#") {
						video.Parts = append(video.Parts, hls.PathJoin(v))
					}
					return false
				},
			),
		)
	}
	return video, nil
}

// resolution .
type resolution struct {
	data map[string]string
}

// newResolution .
func newResolution() *resolution {
	return &resolution{data: make(map[string]string)}
}

// add .
func (r *resolution) add(v string) () {
	if str := cpResolution.FinsSubStringBeforePrefix(v, "hls"); len(str) != 0 {
		r.data["hls-"+str] = v
	}
}

// best .
func (r *resolution) best() string {
	for i := range tendency {
		if v, found := r.data[tendency[i]]; found {
			return v
		}
	}
	return ""
}
