package fail

import (
	"context"
	"github.com/cevaris/status/logging"
	"github.com/cevaris/status/report"
)

// PanicReport throws panic intentionally to confirm we do not crash the runner/scheduler
func PanicReport(name string) (report.ApiReport, error) {
	logger := logging.Logger()
	reportLogger := report.NewLogger(logger)
	ctx := context.Background()

	reportLogger.Debug(ctx,"starting", name)

	panic("EXPECTED: panic test threw a panic")
}