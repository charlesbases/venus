package tools

import "os"

// CreateFile .
func CreateFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

// MkdirAll .
func MkdirAll(name string) error {
	return os.MkdirAll(name, 0755)
}
