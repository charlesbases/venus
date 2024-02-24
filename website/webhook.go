package website

import (
	"github.com/pkg/errors"

	"github.com/charlesbases/venus/sender"
	"github.com/charlesbases/venus/types"
)

var (
	// ErrLinkType .
	ErrLinkType = errors.New("unknown link type")
)

// WebHookFunc .
type WebHookFunc func(link types.LinkPath) WebHook

// Header 视频信息
type Header struct {
	// VideoID 视频 ID
	VideoID string
	// Title 视频标题
	Title string
	// LinkPath 视频链接
	LinkPath types.LinkPath
}

// Video .
type Video struct {
	*Header

	Metadata []sender.Option

	// 视频分片链接列表
	Parts []types.LinkPath
}

// UserInfor 用户信息
type UserInfor struct {
	// ID 用户 ID
	ID string
	// Root 视频网站
	Root string
	// Videos 视频列表
	Videos []*Header
}

// Title .
func (u UserInfor) Title() string {
	if len(u.ID) != 0 {
		return u.Root + "/" + u.ID
	}
	return u.Root
}

// WebHook .
type WebHook interface {
	// Sender .
	Sender() sender.Sender
	// UserInfor 获取用户信息
	UserInfor() (*UserInfor, error)
	// ParseHeader 解析视频信息
	ParseHeader(h *Header) (*Video, error)
}

// Resolution .
type Resolution interface {
	// Add 添加视频分辨率
	Add(v string)
	// Best 最优分辨率
	Best() string
}
