package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/charlesbases/venus/types"
)

const (
	// UserAgentGooglebot 谷歌搜索引擎
	UserAgentGooglebot = "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36"
	// UserAgentGoogleChrome 谷歌浏览器
	UserAgentGoogleChrome = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
)

// HTTPSender use httpSender
var HTTPSender Sender = &httpSender{
	client: &http.Client{
		Timeout:   30 * time.Minute,
		Transport: http.DefaultTransport,
	},
}

var CommandSender Sender = new(commandSender)

// Sender .
type Sender interface {
	// SendRequest 发送请求
	SendRequest(url string, reader Handler, opts ...Option) error
}

// Option .
type Option interface {
	apply(o *options)
}

// options .
type options struct {
	method string
	header map[string]string
	body   io.Reader
}

// newOptions .
func newOptions(opts ...Option) *options {
	o := &options{
		method: http.MethodGet,
		header: map[string]string{
			"User-Agent": UserAgentGooglebot,
		},
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

type method string

// apply .
func (m method) apply(o *options) () {
	o.method = string(m)
}

// Method .
func Method(m string) Option {
	return method(m)
}

type header map[string]string

// apply .
func (h header) apply(o *options) () {
	for key := range h {
		o.header[key] = h[key]
	}
}

// Header .
func Header(h map[string]string) Option {
	return header(h)
}

type params map[string]interface{}

// apply .
func (p params) apply(o *options) () {
	data, _ := json.Marshal(&p)
	o.body = bytes.NewReader(data)
}

// Params .
func Params(p map[string]interface{}) Option {
	return params(p)
}

var _ Sender = (*httpSender)(nil)

// httpSender .
type httpSender struct {
	client *http.Client
}

// SendRequest 发送请求
func (h httpSender) SendRequest(url string, handler Handler, opts ...Option) error {
	o := newOptions(opts...)

	req, err := http.NewRequest(o.method, url, o.body)
	if err != nil {
		return types.NewSenderError(url, err)
	}

	for key := range o.header {
		req.Header.Add(key, o.header[key])
	}

	rsp, err := h.client.Do(req)
	if err != nil {
		return types.NewSenderError(url, err)
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode / 100 {
	case 2:
		if err := handler(rsp.Body); err != nil {
			return types.NewSenderError(url, err)
		}
		return nil
	default:
		return types.NewSenderError(url, types.StatusCode(rsp.StatusCode))
	}
}

var _ Sender = (*commandSender)(nil)

// commandSender .
type commandSender struct{}

// SendRequest 发送请求
func (c commandSender) SendRequest(url string, handler Handler, opts ...Option) error {
	o := newOptions(opts...)

	var args = make([]string, 0, 4)

	// method
	args = append(args, "-X", o.method)

	// header
	for key := range o.header {
		args = append(args, "-H", fmt.Sprintf(`%s: %s`, key, o.header[key]))
	}

	// body TODO

	// URL
	args = append(args, url)

	output, err := exec.Command("curl", args...).Output()
	if err != nil {
		return types.NewSenderError(url, err)
	}

	if err := handler(bytes.NewReader(output)); err != nil {
		return types.NewSenderError(url, err)
	}
	return nil
}
