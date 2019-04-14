package fail

import (
	"context"
	"github.com/cevaris/status/report"
)

// PanicReport throws panic intentionally to confirm we do not crash the runner/scheduler
func PanicReport(ctx context.Context, r report.Request) (report.ApiReport, error) {
	reportLogger := r.ReportLogger

	reportLogger.Debug(ctx, "starting", r.Name)

	panic("EXPECTED: panic test threw a panic")
}
