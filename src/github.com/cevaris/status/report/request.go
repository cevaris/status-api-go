package report

import (
	"cloud.google.com/go/datastore"
	"fmt"
	"github.com/cevaris/timber"
	"time"
)

const (
	KindApiReportMin = "ApiReportMin"
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

func (r *Request) Key() *datastore.Key {
	return datastore.NameKey(
		KindApiReportMin,
		fmt.Sprintf("%s:%d", r.Name, r.TimeMinute.Unix()),
		nil,
	)
}
