package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Log main logger structure
// cool logger implementation can be found here https://github.com/go-pkgz/lgr/blob/master/logger.go
type Log struct {
	stdout, stderr io.Writer
	now            func() time.Time
}

//New creates logger
func New() *Log {
	return &Log{
		stdout: os.Stdout,
		stderr: os.Stderr,
		now:    time.Now,
	}
}

//Logf write a log
func (logger *Log) Logf(line string, args ...interface{}) {
	msg := logger.now().Format("2006/01/02 15:04:05.000") + " " + fmt.Sprintf(line, args...)
	logger.stdout.Write([]byte(msg))
}
