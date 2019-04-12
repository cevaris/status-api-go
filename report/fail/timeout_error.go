package fail

import (
	"github.com/cevaris/status/logging"
	"github.com/cevaris/status/report"
	"time"
)

func TimeoutErrorReport(name string) (report.ApiReport, error) {
	apiReport := report.NewReport(name)
	logger := logging.Logger()
	reportLogger := report.NewLogger(logger)
	now := time.Now().UTC()
	ctx, cancel := report.NewContext()
	defer cancel()

	reportLogger.Info(ctx, "starting", name)

	select {
	// wait longer than max runner time
	case <-time.After(1 * time.Hour):
		reportLogger.Error(ctx, name, "is broken")
	case <-ctx.Done():
		reportLogger.Info(ctx, name, "is being handled correctly")
		return report.NewError(name, reportLogger), ctx.Err()
	}

	later := time.Now().UTC()

	apiReport.LatencyMS = later.Sub(now).Nanoseconds() / int64(time.Millisecond)
	apiReport.Report = reportLogger.Collect()
	apiReport.ReportState = report.Pass // not supposed to reach here

	return apiReport, nil
}
