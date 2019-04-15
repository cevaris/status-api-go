package report

import (
	"github.com/cevaris/timber"
	"time"
)

type Request struct {
	Name         string
	Logger       timber.Logger
	ReportLogger *Logger
	TimeMinute   time.Time
}

// NewRequest creates a logger that should only be used per request!!!
// so we dont overflow the internal buffer
func NewRequest(logger timber.Logger, name string) Request {
	reportLogger := NewLogger(logger)

	return Request{
		Name:         name,
		Logger:       logger,
		ReportLogger: reportLogger,
		TimeMinute:   NowUTCMinute(),
	}
}
