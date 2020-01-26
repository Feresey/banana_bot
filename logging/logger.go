package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	parts string
	Quite bool
}

var (
	mu      = &sync.Mutex{}
	LogFile io.Writer
)

func init() {
	_ = os.Mkdir("log", 0777)
	file, err := os.OpenFile("log/"+time.Now().Format("2006-01-02")+".log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	LogFile = file
}

func NewLogger(parts ...string) *Logger {
	return &Logger{
		parts: strings.Join(parts, ">>") + ">>",
	}
}

func (l *Logger) Child(part ...string) *Logger {
	return &Logger{
		parts: l.parts + strings.Join(part, ">>") + ">>",
	}
}

func (l *Logger) Info(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "INFO:" + fmt.Sprintln(args...))
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Infof(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "INFO:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Warn(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "WARNING:" + fmt.Sprintln(args...))
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "WARNING:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Error(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "ERROR:" + fmt.Sprintln(args...))
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "ERROR:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Fatal(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "FATAL:" + fmt.Sprintln(args...))
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
	os.Exit(1)
}
func (l *Logger) Fatalf(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "FATAL:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = LogFile.Write(msg)
	_, _ = os.Stdout.Write(msg)
	os.Exit(1)
}
func (l *Logger) Panic(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "PANIC:" + fmt.Sprintln(args...))
	_, _ = LogFile.Write(msg)
	panic(string(msg))
}
func (l *Logger) Panicf(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "PANIC:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = LogFile.Write(msg)
	panic(string(msg))
}
