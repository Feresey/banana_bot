package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type Logger struct {
	parts string
	Quite bool
}

var (
	mu = &sync.Mutex{}
)

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
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Infof(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "INFO:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Warn(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "WARNING:" + fmt.Sprintln(args...))
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "WARNING:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Error(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "ERROR:" + fmt.Sprintln(args...))
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "ERROR:" + fmt.Sprintf(format, args...) + "\n")
	_, _ = os.Stdout.Write(msg)
}
func (l *Logger) Fatal(args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "FATAL:" + fmt.Sprintln(args...))
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
	panic(string(msg))
}
func (l *Logger) Panicf(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if l.Quite {
		return
	}
	msg := []byte(l.parts + "PANIC:" + fmt.Sprintf(format, args...) + "\n")
	panic(string(msg))
}
