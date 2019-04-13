package fail

import (
	"context"
	"github.com/cevaris/status/logging"
	"github.com/cevaris/status/report"
)

// PanicReport throws panic intentionally to confirm we do not crash the runner/scheduler
func PanicReport(ctx context.Context, name string) (report.ApiReport, error) {
	logger := logging.FileLogger(name)
	reportLogger := report.NewLogger(logger)

	reportLogger.Debug(ctx,"starting", name)

	panic("EXPECTED: panic test threw a panic")
}