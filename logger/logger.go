package logger

import (
	"fmt"
	"os"
	"sync"
)

const stream = "error.log"

var file *os.File

var lock sync.RWMutex

// Create .
func Create() error {
	f, err := os.OpenFile(stream, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	file = f
	return nil
}

// Error .
func Error(v interface{}) () {
	lock.Lock()
	file.WriteString(fmt.Sprintf(`%v`, v))
	lock.Unlock()
}

// Errorf .
func Errorf(format string, v ...interface{}) {
	lock.Lock()
	file.WriteString(fmt.Sprintf(format, v...))
	lock.Unlock()
}

// Close .
func Close() {
	if file != nil {
		file.Sync()
		file.Close()
	}
}
