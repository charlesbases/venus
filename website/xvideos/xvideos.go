package xvideos

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/charlesbases/venus/website"
)

var _ website.WebHook = (*xvideos)(nil)

const root website.Link = "https://www.xvideos.com"

var resolution = func(v string) int {
	prefixs := []string{"hls-1080p"}
	for idx, prefix := range prefixs {
		if strings.HasPrefix(v, prefix) {
			return len(prefixs) - idx
		}
	}
	return -1
}

// xvideos .
type xvideos struct {
	src  website.Link
	user *website.UserInfor
}

// UserInfor 用户视频信息
func (x *xvideos) UserInfor() (*website.UserInfor, error) {
	// user
	if compileIsUserLink.MatchString(x.src.String()) {
		return x.parseUserInforFromUserLink()
	}

	// video
	if compileIsVideoLink.MatchString(x.src.String()) {
		return x.parseUserInforFromVideoLink()
	}

	// unknown
	return nil, errors.Errorf("%s: unknown link type", x.src)
}

// parseUserInforFromUserLink .
func (x *xvideos) parseUserInforFromUserLink() (*website.UserInfor, error) {
	x.user.ID = compileParseUserIDFromLink.FindSubString(x.src.String())

	// 获取视频列表
	var page int
	for {
		var resp = new(UserVideosResponse)

		// 格式错误
		if err := x.src.Joins("videos", "new", strconv.Itoa(page)).Fetch(website.Unmarshal(resp)); err != nil {
			return nil, err
		}

		// 页码错误
		if len(resp.Videos) == 0 {
			break
		}

		for _, video := range resp.Videos {
			if suffix := complieParseVideoIDFromUserHomePage.FindSubString(video.U); len(suffix) != 0 {
				l := root.Joins("video" + suffix)

				x.user.Videos = append(
					x.user.Videos,
					&website.Header{
						ID:   compileParseVideoIDFromLink.FindSubString(l.String()),
						Link: l,
					},
				)
			}
		}
		page++
	}
	return x.user, nil
}

// UserVideosResponse 艺术家主页
type UserVideosResponse struct {
	Videos []*struct {
		U string `json:"u"`
	} `json:"videos"`
}

// parseUserInforFromVideoLink .
func (x *xvideos) parseUserInforFromVideoLink() (*website.UserInfor, error) {
	x.user.Videos = append(
		x.user.Videos, &website.Header{
			ID:   compileParseVideoIDFromLink.FindSubString(x.src.String()),
			Link: x.src,
		},
	)
	return x.user, nil
}

// ParseVideo 根据视频网页链接，获取下载地址
func (x *xvideos) ParseVideo(h *website.Header) (*website.Video, error) {
	video := &website.Video{Header: h}

	// parse hls link
	if err := h.Link.Fetch(
		website.ReadLine(
			func(line string) (isBreak bool) {
				if len(video.HLink) == 0 {
					video.HLink = website.Link(complieParseHLinkFromUserHomePage.FindSubString(line))
				}
				if len(x.user.ID) == 0 {
					x.user.ID = compileParseUserIDFromUserHomePage.FindSubString(line)
				}
				return len(x.user.ID) != 0 && len(video.HLink) != 0
			},
		),
	); err != nil {
		return nil, err
	}

	if len(video.HLink) == 0 {
		return nil, errors.New("not found")
	}

	return x.parseVideoParts(video), nil
}

// parseVideoParts .
func (x *xvideos) parseVideoParts(video *website.Video) *website.Video {
	var rst = website.NewResolutionRule(resolution)

	// 分辨率
	video.HLink.Joins("hls.m3u8").Fetch(
		website.ReadLine(
			func(line string) (isBreak bool) {
				if !strings.HasPrefix(line, "#") {
					rst.Add(line)
				}
				return false
			},
		),
	)

	// 	视频下载列表
	video.Parts = make([]website.Link, 0)
	video.HLink.Joins(rst.Best()).Fetch(
		website.ReadLine(
			func(line string) (isBreak bool) {
				if !strings.HasPrefix(line, "#") {
					video.Parts = append(video.Parts, video.HLink.Joins(line))
				}
				return false
			},
		),
	)

	return video
}

// New .
func New(src website.Link) website.WebHook {
	return &xvideos{
		src: src,
		user: &website.UserInfor{
			Root: "xvideos.com",
		},
	}
}
