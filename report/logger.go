package report

import (
	"context"
	"fmt"
	"github.com/cevaris/timber"
	"strings"
)

func NewLogger(logger timber.Logger) *Logger {
	return &Logger{underlying: logger, logs: make([]string, 0)}
}

type Logger struct {
	underlying timber.Logger
	logs       []string
}

func (l *Logger) Info(ctx context.Context, m ...interface{}) {
	appendEntries(l, m)
	l.underlying.Info(ctx, m...)
}

func (l *Logger) Error(ctx context.Context, m ...interface{}) {
	appendEntries(l, m)
	l.underlying.Error(ctx, m...)
}

func (l *Logger) Debug(ctx context.Context, m ...interface{}) {
	appendEntries(l, m)
	l.underlying.Debug(ctx, m...)
}

// Returns collected logs
func (l *Logger) Collect() []string {
	return l.logs
}

func appendEntries(l *Logger, m []interface{}) {
	// convert message params to strings
	line := make([]string, len(m))
	for _, a := range m {
		line = append(line, fmt.Sprintf("%+v", a))
	}
	// add as a single line entry
	l.logs = append(l.logs, strings.Join(line[:], " "))
}
