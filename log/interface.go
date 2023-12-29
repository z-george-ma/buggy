package log

type Logger interface {
	Debug() LogEntry
	Info() LogEntry
	Warn() LogEntry
	Err() LogEntry
	Fatal() LogEntry
	With() LogContext
}

type LogContext interface {
	Unit(service string) LogContext
	Caller(skip int) LogContext
	Value(key string, value any) LogContext
	Logger() Logger
}

type LogEntry interface {
	Caller(skip int) LogEntry
	Value(key string, value any) LogEntry
	Msg(string)
	Error(skip int, err error)
}
