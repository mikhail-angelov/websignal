package logger

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoggerNoDbg(t *testing.T) {
	rout := bytes.NewBuffer([]byte{})
	l := New()
	l.stdout = rout
	l.now = func() time.Time { return time.Date(2019, 10, 20, 15, 37, 0, 0, time.Local) }
	l.Logf("test %d", 42)
	assert.Equal(t, "2019/10/20 15:37:00.000 test 42", rout.String())
}
