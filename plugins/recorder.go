package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/charlesbases/venus/tools"
)

const filename = "./vencus.json"

// recorder .
type recorder interface {
	Write(domain string, id string, opts ...option)
	Close() error
}

var _ recorder = (*fileRecorder)(nil)

type (
	// fileRecorder .
	fileRecorder struct {
		name  string
		cache map[string]struct{}
		lock  sync.RWMutex

		List []*WebHome
	}

	// WebHome .
	WebHome struct {
		Root   string         `json:"root"`
		Count  int            `json:"count"`
		Videos []*VideoHeader `json:"videos"`
	}

	// VideoHeader .
	VideoHeader struct {
		ID    string `json:"id"`
		User  string `json:"user,omitempty"`
		Date  string `json:"date,omitempty"`
		Link  string `json:"link,omitempty"`
		Title string `json:"title,omitempty"`
	}
)

// option .
type option func(v *VideoHeader) ()

// recordDate .
func recordDate(t time.Time) option {
	return func(v *VideoHeader) {
		v.Date = t.Format(time.DateTime)
	}
}

// recordUser .
func recordUser(name string) option {
	return func(v *VideoHeader) {
		v.User = name
	}
}

// recordLink .
func recordLink(link string) option {
	return func(v *VideoHeader) {
		v.Link = link
	}
}

// recordTitle .
func recordTitle(title string) option {
	return func(v *VideoHeader) {
		v.Title = title
	}
}

// Write 记录视频信息
func (r *fileRecorder) Write(domain string, id string, opts ...option) {
	r.lock.Lock()
	defer r.lock.Unlock()

	var idstr = domain + "/" + id
	if _, found := r.cache[idstr]; found {
		return
	}

	var video = &VideoHeader{ID: id, Date: time.Now().Format(time.DateTime)}
	for _, opt := range opts {
		opt(video)
	}

	var webhome *WebHome
	for _, v := range r.List {
		if v.Root == domain {
			webhome = v
			break
		}
	}
	if webhome == nil {
		webhome = &WebHome{
			Root: domain,
		}
		r.List = append(r.List, webhome)
	}
	webhome.Videos = append(webhome.Videos, video)

	r.cache[idstr] = struct{}{}
}

// Close .
func (r *fileRecorder) Close() error {
	r.sort()
	file, err := tools.CreateFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(&r.List, "", "  ")
	if err != nil {
		return err
	}

	file.Write(data)
	return nil
}

// sort .
func (r *fileRecorder) sort() () {
	for _, v := range r.List {
		v.Count = len(v.Videos)

		sort.Slice(
			v.Videos, func(i, j int) bool {
				if v.Videos[i].User == v.Videos[j].User {
					return v.Videos[i].ID > v.Videos[j].ID
				}
				return v.Videos[i].User < v.Videos[j].User
			},
		)
	}

	sort.Slice(
		r.List, func(i, j int) bool {
			return r.List[i].Root < r.List[j].Root
		},
	)
}

// newRecorder .
func newRecorder() (recorder, error) {
	absname, _ := filepath.Abs(filename)

	var rec = &fileRecorder{name: absname, cache: make(map[string]struct{})}

	data, _ := os.ReadFile(absname)
	if len(data) != 0 {
		if err := json.Unmarshal(data, &rec.List); err != nil {
			return nil, err
		}
		for _, val := range rec.List {
			rec.cache[val.Root] = struct{}{}
		}
	}

	return rec, nil
}
