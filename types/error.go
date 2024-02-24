package types

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// CustomError .
type CustomError struct {
	domain string
	err    error
}

// Error .
func (e CustomError) Error() string {
	return fmt.Sprintf("%s\n└─── %v\n", e.domain, e.err)
}

// NewCustomError .
func NewCustomError(domain string, err error) error {
	return &CustomError{domain: domain, err: err}
}

// StatusCode 状态码
type StatusCode int

// Error .
func (c StatusCode) Error() string {
	return strconv.Itoa(int(c))
}

// NewSenderError .
func NewSenderError(url string, err error) error {
	return NewCustomError(url, errors.Wrap(err, "HTTP"))
}
