package fail

import (
	"context"
	"github.com/cevaris/status/logging"
	"github.com/cevaris/status/report"
)

func PanicReport(name string) (report.ApiReport, error) {
	logger := logging.Logger()
	reportLogger := report.NewLogger(logger)
	ctx := context.Background()

	reportLogger.Debug(ctx,"starting", name)

	panic("EXPECTED: panic test threw a panic")
}