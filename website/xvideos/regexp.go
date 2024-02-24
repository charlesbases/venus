package xvideos

import "github.com/charlesbases/venus/types"

var (
	// 用户链接
	cpIsUserLink = types.MustCompile(`https://www.xvideos.com/.*/[a-zA-Z0-9_-]+`)
	// 视频链接
	cpIsVideoLink = types.MustCompile(`https://www.xvideos.com/video[0-9]+/_`)
)

// 根据视频链接获取视频 ID
var cpParseVideoIDFromLink = types.MustCompile(`https://www[.]xvideos[.]com/video([0-9]+)`)

var (
	// 视频分辨率
	cpResolution = types.MustCompile(`hls-(.*)-.*`)
	// 根据视频链接获取用户 id
	cpParseUserIDFromLink = types.MustCompile(`https://www.xvideos.com/.*/(.*)`)
	// 根据用户首页的视频列表获取视频 id
	cpParseVideoIDFromUserHomePage = types.MustCompile(`.*/([1-9]+.*)`)
	// 根据用户首页的视频列表获取视频下载链接
	cpParseHLinkFromUserHomePage = types.MustCompile(`html5player[.]setVideoHLS[(]'(.*)/hls.m3u8'[)];`)
	// 根据用户首页的视频列表获取用户 ID
	cpParseUserIDFromUserHomePage = types.MustCompile(`html5player[.]setUploaderName[(]'(.*)'[)];`)
)
