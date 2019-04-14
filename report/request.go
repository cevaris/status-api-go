package report

import (
	"github.com/cevaris/timber"
)

type Request struct {
	Name         string
	Logger       timber.Logger
	ReportLogger *Logger
}

// NewRequest creates a logger that should only be used per request!!!
// so we dont overflow the internal buffer
func NewRequest(logger timber.Logger, name string) Request {
	reportLogger := NewLogger(logger)

	return Request{
		Name:         name,
		Logger:       logger,
		ReportLogger: reportLogger,
	}
}
