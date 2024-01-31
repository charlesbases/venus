package website

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrorCodeUnkown .
	ErrorCodeUnkown errorCode = -1
	// ErrorCodeNotFound 404
	ErrorCodeNotFound errorCode = 404
)

var defaultClient = &http.Client{Timeout: 15 * time.Minute}

var defaultHeader = map[string]string{
	// "Content-Type": "application/json",
	// "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"User-Agent": "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36",
}

type errorCode int

// Error .
func (code errorCode) Error() string {
	return strconv.Itoa(int(code))
}

// Link 链接
type Link string

// String .
func (l Link) String() string {
	return string(l)
}

// Fetch .
func (l Link) Fetch(fn reader, opts ...func(meta *Metadata)) error {
	return fetch(l, fn, opts...)
}

// Joins .
func (l Link) Joins(v ...string) Link {
	var br strings.Builder
	var n = len(l.String())

	for _, val := range v {
		n += len(val)
	}

	br.Grow(n)
	br.WriteString(l.String())

	for _, val := range v {
		br.WriteString("/")
		br.WriteString(strings.TrimPrefix(val, "/"))
	}
	return Link(br.String())
}

// Metadata .
type Metadata struct {
	Method string
	Header map[string]string
	Data   interface{}
	body   io.Reader
}

// fetch .
func fetch(l Link, fn reader, opts ...func(meta *Metadata)) error {
	var meta = &Metadata{Method: http.MethodGet, Header: make(map[string]string, 0)}
	for _, opt := range opts {
		opt(meta)
	}

	if meta.Data != nil {
		data, _ := json.Marshal(meta.Data)
		meta.body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(meta.Method, l.String(), meta.body)
	if err != nil {
		return errors.Wrap(err, "http")
	}

	for key, val := range defaultHeader {
		req.Header.Add(key, val)
	}
	for key, val := range meta.Header {
		req.Header.Add(key, val)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return errors.Wrap(fn(resp.Body), l.String())
	case 404:
		return errors.Wrap(ErrorCodeNotFound, l.String())
	default:
		return errors.Wrap(ErrorCodeUnkown, l.String())
	}
}

type reader func(r io.Reader) error

// ReadLine .
func ReadLine(fn func(line string) (isBreak bool)) reader {
	return func(r io.Reader) error {
		var buf = bufio.NewReader(r)
		for {
			if line, err := buf.ReadString('\n'); err != nil {
				break
			} else {
				if fn(strings.TrimSuffix(line, "\n")) {
					break
				}
			}
		}
		return nil
	}
}

// ReadAll .
func ReadAll(fn func(data string) error) reader {
	return func(r io.Reader) error {
		data, _ := io.ReadAll(r)
		return fn(string(data))
	}
}

// WriteTo .
func WriteTo(w io.Writer) reader {
	return func(r io.Reader) error {
		_, err := io.Copy(w, r)
		return err
	}
}

// Unmarshal .
func Unmarshal(v interface{}) reader {
	return func(r io.Reader) error {
		if data, err := io.ReadAll(r); err != nil {
			return errors.Wrap(err, "http")
		} else {
			return errors.Wrap(json.Unmarshal(data, v), "json unmarshal")
		}
	}
}