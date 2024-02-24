package sender

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// Handler .
type Handler func(r io.Reader) error

// ReadAll .
func ReadAll(fn func(v string) error) Handler {
	return func(r io.Reader) error {
		data, _ := io.ReadAll(r)
		return fn(string(data))
	}
}

// ReadLine .
func ReadLine(fn func(line string) (isBreak bool)) Handler {
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

// WriteTo .
func WriteTo(w io.Writer) Handler {
	return func(r io.Reader) error {
		io.Copy(w, r)
		return nil
	}
}

// Unmarshal .
func Unmarshal(v interface{}) Handler {
	return func(r io.Reader) error {
		if data, err := io.ReadAll(r); err != nil {
			return errors.Wrap(err, "IO")
		} else {
			return errors.Wrap(json.Unmarshal(data, v), "JsonUnmarshal")
		}
	}
}
