package xvideos

import "github.com/charlesbases/venus/regexp"

var (
	// 用户链接
	compileIsUserLink = regexp.New(`https://www.xvideos.com/.*/[a-zA-Z0-9_-]+`)
	// 视频链接
	compileIsVideoLink = regexp.New(`https://www.xvideos.com/video[0-9]+/_`)
)

// 根据视频链接获取视频 ID
var compileParseVideoIDFromLink = regexp.New(`https://www[.]xvideos[.]com/video([0-9]+)`)

var (
	// 根据视频链接获取用户 id
	compileParseUserIDFromLink = regexp.New(`https://www.xvideos.com/.*/(.*)`)
	// 根据用户首页的视频列表获取视频 id
	complieParseVideoIDFromUserHomePage = regexp.New(`.*/([1-9]+.*)`)
	// 根据用户首页的视频列表获取视频下载链接
	complieParseHLinkFromUserHomePage = regexp.New(`html5player[.]setVideoHLS[(]'(.*)/hls.m3u8'[)];`)
	// 根据用户首页的视频列表获取用户 ID
	compileParseUserIDFromUserHomePage = regexp.New(`html5player[.]setUploaderName[(]'(.*)'[)];`)
)
