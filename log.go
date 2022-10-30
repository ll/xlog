package xlog

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

type Logger struct {
	dst       *bufio.Writer
	file      *os.File
	lock      sync.Mutex
	lastFlush time.Time
}

var defaultLogger = &Logger{lastFlush: time.Now()}

func InitLog(path string) {
	var err error
	defaultLogger.file, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}

	defaultLogger.dst = bufio.NewWriter(defaultLogger.file)
}

func Flush() error {
	return defaultLogger.Flush()
}

func (l *Logger) Flush() error {
	if err := l.dst.Flush(); err != nil {
		return err
	}
	return l.file.Sync()
}

func Infof(msg string, args ...interface{}) {
	defaultLogger.Infof(msg, args...)
}

func Warnf(msg string, args ...interface{}) {
	defaultLogger.Warnf(msg, args...)
}

func Errorf(msg string, args ...interface{}) {
	defaultLogger.Errorf(msg, args...)
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	l.logf(levelInfo, msg, args)
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.logf(levelWarn, msg, args)
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.logf(levelError, msg, args)
}

type logLevel int8

const (
	levelDebug logLevel = iota + 1
	levelInfo
	levelWarn
	levelError
)

func (l *Logger) logf(level logLevel, msgStr string, args []interface{}) {
	msg := String2Bytes(msgStr)

	l.lock.Lock()
	defer l.lock.Unlock()

	switch level {
	case levelDebug:
		l.dst.Write([]byte("DEBUG "))
	case levelInfo:
		l.dst.Write([]byte("INFO "))
	case levelWarn:
		l.dst.Write([]byte("WARN "))
	case levelError:
		l.dst.Write([]byte("ERROR "))
	}
	tn := time.Now()
	l.dst.Write([]byte(tn.Format("01-02 15:04:05.999 ")))

	// build formated message
	var i, cnt int
	for j := 0; j < len(msg); j++ {
		// find %v
		if msg[j] != '%' || j+1 == len(msg) || msg[j+1] != 'v' {
			continue
		}
		// write msg before %v
		l.dst.Write([]byte(msg[i:j]))
		i = j + 2
		j++ // for 循环体中还有自增
		// write %v's value
		if cnt < len(args) {
			l.dst.Write(getValue(args[cnt]))
			cnt++
		} else {
			l.dst.Write([]byte("args needed"))
		}
	}
	// write tail
	if i < len(msg) {
		l.dst.Write(msg[i:])
	}
	l.dst.Write([]byte("\n"))

	if tn.After(l.lastFlush.Add(time.Second * 2)) {
		l.Flush()
	}
}

func getValue(arg interface{}) []byte {
	switch a := arg.(type) {
	case int:
		return []byte(strconv.FormatInt(int64(a), 10))
	case int8:
		return []byte(strconv.FormatInt(int64(a), 10))
	case int16:
		return []byte(strconv.FormatInt(int64(a), 10))
	case int32:
		return []byte(strconv.FormatInt(int64(a), 10))
	case int64:
		return []byte(strconv.FormatInt(a, 10))
	case string:
		return String2Bytes(a)
	default:
		return []byte(fmt.Sprintf("(unknown type: %v)", a))
	}
	return nil
}

func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
