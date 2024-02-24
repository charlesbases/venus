package logger

import (
	"fmt"
	"os"

	"github.com/charlesbases/venus/tools"
)

const stream = "error.log"

var log struct {
	f *os.File
	c chan string
}

// Create .
func Create() error {
	f, err := tools.CreateFile(stream)
	if err != nil {
		return err
	}
	log.f = f
	log.c = make(chan string)
	go daemon()
	return nil
}

// daemon .
func daemon() () {
	for {
		select {
		case mess, ok := <-log.c:
			if !ok {
				log.f.Sync()
				log.f.Close()
				return
			}
			log.f.WriteString(mess + "\n")
		}
	}
}

// Error .
func Error(v interface{}) () {
	log.c <- fmt.Sprintf(`%v`, v)
}

// Errorf .
func Errorf(format string, v ...interface{}) () {
	log.c <- fmt.Sprintf(format, v...)
}

// Close .
func Close() () {
	close(log.c)
}
