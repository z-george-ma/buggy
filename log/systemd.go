package log

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"runtime"
	"strings"
	"sync"

	"github.com/z-george-ma/buggy/v2/lib"
)

type SystemdLogger struct {
	conn       net.Conn
	mapBuf     *sync.Pool
	ch         chan map[string]any
	onError    func(error) bool
	loopCancel context.CancelFunc
	loopEnded  chan struct{}
}

type SystemdLogEntry struct {
	logger *SystemdLogger
	dict   map[string]any
}

type SystemdLogContext SystemdLogEntry
type SystemdLogContextLogger SystemdLogContext

func (self *SystemdLogger) loop(ctx context.Context) {
	defer close(self.loopEnded)
	buf := &bytes.Buffer{}

	var m map[string]any
	for {
		select {
		case <-ctx.Done():
			return
		case m = <-self.ch:
		}

		buf.Reset()

		for k, v := range m {
			buf.WriteString(k)

			if s, ok := v.(string); ok {
				buf.WriteString("\n")
				binary.Write(buf, binary.LittleEndian, int64(len(s)))
				buf.WriteString(s)
			} else {
				buf.WriteString("=")
				buf.WriteString(lib.Cast[string](v))
			}

			buf.WriteString("\n")
		}

		clear(m)
		self.mapBuf.Put(m)

		_, err := self.conn.Write(buf.Bytes())

		if err != nil && self.onError(err) {
			return
		}
	}
}

func NewLogger(onError func(error) bool) (ret *SystemdLogger, err error) {
	conn, err := net.Dial("unixgram", "/run/systemd/journal/socket")

	if err != nil {
		return
	}

	ret = &SystemdLogger{
		conn: conn,
		mapBuf: &sync.Pool{
			New: func() any {
				return map[string]any{}
			},
		},
		ch:        make(chan map[string]any, 100),
		loopEnded: make(chan struct{}),
		onError:   onError,
	}

	loopCtx, cancel := context.WithCancel(context.Background())
	ret.loopCancel = cancel
	go ret.loop(loopCtx)

	return
}

func (self *SystemdLogger) Close(ctx context.Context) {
	close(self.ch)
	select {
	case <-ctx.Done():
	case <-self.loopEnded:
	}

	self.conn.Close()
}

func (self *SystemdLogger) createLogger(logLevel int, dict map[string]any) LogEntry {
	dict["PRIORITY"] = logLevel

	return &SystemdLogEntry{
		logger: self,
		dict:   dict,
	}
}

func (self *SystemdLogger) Debug() LogEntry {
	return self.createLogger(7, self.mapBuf.Get().(map[string]any))
}

func (self *SystemdLogger) Info() LogEntry {
	return self.createLogger(6, self.mapBuf.Get().(map[string]any))
}

func (self *SystemdLogger) Warn() LogEntry {
	return self.createLogger(4, self.mapBuf.Get().(map[string]any))
}

func (self *SystemdLogger) Err() LogEntry {
	return self.createLogger(3, self.mapBuf.Get().(map[string]any))
}

func (self *SystemdLogger) Fatal() LogEntry {
	return self.createLogger(2, self.mapBuf.Get().(map[string]any))
}

func (self *SystemdLogger) With() LogContext {
	return &SystemdLogContext{
		logger: self,
		dict:   self.mapBuf.Get().(map[string]any),
	}
}

func (self *SystemdLogContext) Unit(service string) LogContext {
	self.dict["UNIT"] = service + ".service"
	return self
}

func (self *SystemdLogContext) Caller(skip int) LogContext {
	_, file, line, ok := runtime.Caller(skip + 1)
	if ok {
		self.dict["CODE_FILE"] = file
		self.dict["CODE_LINE"] = line
	}

	return self
}

func (self *SystemdLogContext) Value(key string, value any) LogContext {
	self.dict[strings.ToUpper(key)] = value
	return self
}

func (self *SystemdLogContext) Logger() Logger {
	return &SystemdLogContextLogger{
		logger: self.logger,
		dict:   self.dict,
	}
}

func (self *SystemdLogContextLogger) Debug() LogEntry {
	return self.logger.createLogger(7, self.dict)
}

func (self *SystemdLogContextLogger) Info() LogEntry {
	return self.logger.createLogger(6, self.dict)
}

func (self *SystemdLogContextLogger) Warn() LogEntry {
	return self.logger.createLogger(4, self.dict)
}

func (self *SystemdLogContextLogger) Err() LogEntry {
	return self.logger.createLogger(3, self.dict)
}

func (self *SystemdLogContextLogger) Fatal() LogEntry {
	return self.logger.createLogger(2, self.dict)
}

func (self *SystemdLogContextLogger) With() LogContext {
	return &SystemdLogContext{
		logger: self.logger,
		dict:   self.dict,
	}
}

func (self *SystemdLogEntry) Caller(skip int) LogEntry {
	_, file, line, ok := runtime.Caller(skip + 1)
	if ok {
		self.dict["CODE_FILE"] = file
		self.dict["CODE_LINE"] = line
	}

	return self
}

func (self *SystemdLogEntry) Value(key string, value any) LogEntry {
	self.dict[strings.ToUpper(key)] = value
	return self
}

func (self *SystemdLogEntry) Error(skip int, err error) {
	self.Caller(skip + 1)
	self.dict["MESSAGE"] = err.Error()

	self.logger.ch <- self.dict
}

func (self *SystemdLogEntry) Msg(msg string) {
	self.dict["MESSAGE"] = msg

	self.logger.ch <- self.dict
}
