package fail

import (
	"context"
	"github.com/cevaris/status/report"
	"time"
)

func TimeoutErrorReport(ctx context.Context, r *report.Request) (report.ApiReport, error) {
	reportLogger := r.ReportLogger

	reportLogger.Info(ctx, "starting", r.Name)
	time.Sleep(1 * time.Hour)

	// should force timeout of report run
	var apiReport report.ApiReport
	return apiReport, nil
}
