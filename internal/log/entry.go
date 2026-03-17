package log

import (
	"fmt"
	"time"
)

type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelDebug LogLevel = "DEBUG"
)

type LogEntry struct {
	Timestamp time.Time
	Source    string
	Level    LogLevel
	Message  string
}

func (e LogEntry) String() string {
	return fmt.Sprintf("%s | %-8s | %-5s | %s",
		e.Timestamp.Format("15:04:05"),
		e.Source,
		e.Level,
		e.Message,
	)
}
