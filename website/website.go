package website

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/charlesbases/progressbar"
	"github.com/charlesbases/salmon"

	"github.com/charlesbases/venus/logger"
)

// format 视频文件格式
const format = "mp4"

// root 视频资源文件夹
var root = func() string {
	abs, _ := filepath.Abs("resource")
	return abs
}()

// lock for resource
var lock sync.RWMutex

// 已下载的资源列表
var resource = func() map[string]struct{} {
	var res = make(map[string]struct{}, 0)
	filepath.Walk(
		root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return filepath.SkipAll
			}

			if !info.IsDir() && strings.HasSuffix(path, format) {
				res[pathSplit(path)] = struct{}{}
			}
			return nil
		},
	)
	return res
}()

// 并发下载数
var concurrent = 10

// SetConcurrent .
func SetConcurrent(c int) {
	if c > 0 {
		concurrent = c
	}
}

// pathJoin .
func pathJoin(v ...string) string {
	return strings.Join(v, "/")
}

// pathSplit .
func pathSplit(v string) string {
	v = strings.TrimPrefix(v, root)
	if len(v) > 0 {
		s := strings.Split(v, string(v[0]))
		if len(s) > 2 {
			return pathJoin(s[1], s[len(s)-1])
		}
	}
	return ""
}

// ResolutionRule 分辨率排序规则
type ResolutionRule struct {
	// 分辨率列表
	list []string
	// 分辨率排序规则
	rule func(v string) int
}

// NewResolutionRule .
func NewResolutionRule(rule func(v string) int) *ResolutionRule {
	return &ResolutionRule{rule: rule, list: make([]string, 0, 8)}
}

// Add .
func (r *ResolutionRule) Add(v string) {
	r.list = append(r.list, v)
}

// Best .
func (r *ResolutionRule) Best() string {
	if len(r.list) != 0 {
		// 根据 rule 规则对 list 排序
		sort.Slice(
			r.list, func(i, j int) bool {
				return r.rule(r.list[i]) > r.rule(r.list[j])
			},
		)
		return r.list[0]
	}
	return ""
}

// Header .
type Header struct {
	// 视频 ID
	ID string
	// 视频链接
	Link Link
}

// throwError .
func (h Header) throwError(err error) () {
	logger.Errorf("%s: %v", h.Link, err)
}

// Video .
type Video struct {
	*Header

	// 下载的视频文件路径
	store string
	// hls 链接
	HLink Link
	// 视频分片链接列表
	Parts []Link
}

// remove 若视频文件下载失败，则清理视频文件
func (v Video) remove() () {
	os.Remove(v.store)
}

// commit 视频文件下载成功后，保存文件
func (v Video) commit() () {
	store := v.store + "." + format
	os.Rename(v.store, store)

	lock.Lock()
	resource[pathSplit(store)] = struct{}{}
	lock.Unlock()
}

// UserInfor .
type UserInfor struct {
	// 用户
	ID string
	// 视频网站
	Root Link
	// 下载文件夹
	store string
	// 视频列表
	Videos []*Header
}

// mkdirall .
func (u *UserInfor) mkdirall() error {
	u.store = filepath.Join(root, u.Root.String(), u.ID)
	return os.MkdirAll(u.store, 0755)
}

// exists 视频文件是否已下载
func (u *UserInfor) exists(h *Header) bool {
	lock.RLock()
	defer lock.RUnlock()

	_, found := resource[pathJoin(u.Root.String(), h.ID+"."+format)]
	return found
}

// create .
func (u *UserInfor) create(v *Video) (*os.File, error) {
	v.store = filepath.Join(u.store, v.ID)
	return os.Create(v.store)
}

// WebHook .
type WebHook interface {
	// UserInfor 用户视频信息
	UserInfor() (*UserInfor, error)
	// ParseVideo 根据视频网页链接，获取下载地址
	ParseVideo(h *Header) (*Video, error)
}

// Crawler .
func Crawler(hooks ...WebHook) error {
	bar := progressbar.NewReader()

	// 并发下载
	pool, err := salmon.NewPool(concurrent, iterator)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(hooks))

	for i := range hooks {
		go func(hook WebHook) {
			defer wg.Done()

			// 获取视频用户信息
			user, err := hook.UserInfor()
			if err != nil {
				logger.Error(err)
				return
			}

			// 创建视频文件夹
			if err := user.mkdirall(); err != nil {
				logger.Error(err)
				return
			}

			pb := bar.NewProgress(user.Root.Joins(user.ID).String(), uint(len(user.Videos)))

			for _, header := range user.Videos {
				pool.Invoke(
					mess{
						hook:   hook,
						user:   user,
						header: header,
						pb:     pb,
					},
				)
			}
		}(hooks[i])
	}

	wg.Wait()
	pool.Wait()
	bar.Close()
	return nil
}

// mess .
type mess struct {
	hook   WebHook
	user   *UserInfor
	header *Header

	pb *progressbar.Progress
}

// iterator .
func iterator(v interface{}, stop func()) {
	m, ok := v.(mess)
	if !ok {
		return
	}

	hook, user, header := m.hook, m.user, m.header

	if user.exists(header) {
		goto finished
	}

	if video, err := hook.ParseVideo(header); err != nil {
		header.throwError(err)
	} else {
		if file, err := user.create(video); err != nil {
			header.throwError(err)
		} else {
			var err error
			// 视频分片下载
			for _, part := range video.Parts {
				if err = part.Fetch(WriteTo(file)); err != nil {
					header.throwError(err)
					break
				}
			}
			file.Close()

			if err != nil {
				video.remove()
			} else {
				video.commit()
			}
		}
	}

finished:
	m.pb.Incr(1)
}
