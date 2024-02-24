package types

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlesbases/venus/tools"
)

// format 视频文件格式
const format = ".mp4"

// Storage 视频资源文件夹
var Storage = rootPath("resource")

// LinkPath .
type LinkPath interface {
	PathJoin(v ...string) LinkPath
	String() string
}

var _ LinkPath = (*linkPath)(nil)

// linkPath .
type linkPath string

// Link .
func Link(v string) LinkPath {
	return linkPath(v)
}

// PathJoin 路径拼接
func (l linkPath) PathJoin(v ...string) LinkPath {
	if len(v) == 0 {
		return l
	}

	var b strings.Builder

	b.WriteString(l.String())
	for i := range v {
		if !strings.HasSuffix(b.String(), "/") && !strings.HasPrefix(v[i], "/") {
			b.WriteString("/")
		}
		b.WriteString(v[i])
	}
	return Link(b.String())
}

// String .
func (l linkPath) String() string {
	return string(l)
}

// DirPath 文件夹路径
type DirPath interface {
	PathJoin(v ...string) DirPath
	FileIsExists(name string) bool
	CreateFile(name string) FilePath
	MkdirAll() DirPath
	String() string
	dirname(depth int) string
}

var _ DirPath = (*dirPath)(nil)

type dirPath string

// MkdirAll .
func (d dirPath) MkdirAll() DirPath {
	tools.MkdirAll(d.String())
	return d
}

// rootPath .
func rootPath(root string) DirPath {
	p, _ := filepath.Abs(root)
	return dirPath(p)
}

// PathJoin .
func (d dirPath) PathJoin(v ...string) DirPath {
	return dirPath(filepath.Join(append([]string{d.String()}, v...)...))
}

// FileIsExists .
func (d dirPath) FileIsExists(name string) bool {
	_, err := os.Stat(d.PathJoin(name + format).String())
	return err == nil
}

// CreateFile .
func (d dirPath) CreateFile(name string) FilePath {
	f, _ := tools.CreateFile(d.PathJoin(name).String())
	return &file{f}
}

// String .
func (d dirPath) String() string {
	return string(d)
}

// dirname 上层文件夹名称
func (d dirPath) dirname(depth int) string {
	var dir = d.String()
	for i := 0; i < depth; i++ {
		dir = filepath.Dir(dir)
	}
	return filepath.Base(dir)
}

// FilePath 文件路径
type FilePath interface {
	Write(fn func(w io.Writer) error) error
}

var _ FilePath = (*file)(nil)

// file .
type file struct {
	*os.File
}

// Writer .
func (f file) Write(fn func(w io.Writer) error) error {
	err := fn(f.File)
	f.Close()

	if err != nil {
		os.Remove(f.Name())
		return err
	}

	os.Rename(f.Name(), f.Name()+format)
	return nil
}
