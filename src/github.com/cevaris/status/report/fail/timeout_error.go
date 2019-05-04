package fail

import (
	"context"
	"github.com/cevaris/status/report"
	"time"
)

func TimeoutErrorReport(ctx context.Context, r report.Request) (report.ApiReport, error) {
	reportLogger := r.ReportLogger

	reportLogger.Info(ctx, "starting", r.Name)
	time.Sleep(1 * time.Hour)

	panic(r.Name + " failure report is broken; expected timeout error")
}
