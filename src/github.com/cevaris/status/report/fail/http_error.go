package fail

import (
	"context"
	"github.com/cevaris/status/report"
	"net/http"
)

func HTTPErrorReport(ctx context.Context, r report.Request) (report.ApiReport, error) {
	reportLogger := r.ReportLogger

	reportLogger.Debug(ctx, "starting", r.Name)

	_, err := http.Get("really broken api hostname")
	if err != nil {
		reportLogger.Error(ctx, "EXPECTED: failed to get", err)
		return report.NewApiReportErr(r), err
	}

	panic(r.Name + " failure report is broken; expected http error")
}
