package fail

import (
	"context"
	"github.com/cevaris/status/logging"
	"github.com/cevaris/status/report"
	"net/http"
)

func HTTPErrorReport(name string) (report.ApiReport, error) {
	logger := logging.Logger()
	reportLogger := report.NewLogger(logger)
	ctx := context.Background()

	reportLogger.Debug(ctx,"starting", name)

	_, err := http.Get("http://no-such-api.com")
	if err != nil {
		reportLogger.Error(ctx, "EXPECTED: failed to get", err)
		return report.NewError(name, reportLogger), err
	}

	var apiReport report.ApiReport
	return apiReport, err
}