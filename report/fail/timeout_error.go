package fail

import (
	"context"
	"github.com/cevaris/status/logging"
	"github.com/cevaris/status/report"
	"time"
)

func TimeoutErrorReport(ctx context.Context, name string) (report.ApiReport, error) {
	logger := logging.FileLogger(name)
	reportLogger := report.NewLogger(logger)

	reportLogger.Info(ctx, "starting", name)
	time.Sleep(1 * time.Hour)

	// should force timeout of report run
	var apiReport report.ApiReport
	return apiReport, nil
}
