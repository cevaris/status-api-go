package fileio

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cevaris/status/report"
)

// WriteTextReport reports on writing a message to https://www.file.io
func WriteTextReport(ctx context.Context, r report.Request) (report.ApiReport, error) {
	reportLogger := r.ReportLogger

	reportLogger.Debug(ctx, "starting test:", r.Name)

	data := url.Values{}
	data.Add("text", fmt.Sprintf("secret number %d", time.Now().UTC().Unix()))
	reportLogger.Debug(ctx, "posting: ", data)

	now := time.Now().UTC()
	resp, err := http.PostForm("https://file.io", data)
	if err != nil {
		reportLogger.Error(ctx, "post failed: "+err.Error())
		return report.NewApiReportErr(r), err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			reportLogger.Error(ctx, "failed to close response reader:", err)
		}
	}()
	later := time.Now().UTC()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reportLogger.Error(ctx, "failed reading response body: "+err.Error())
		return report.NewApiReportErr(r), err
	}
	reportLogger.Debug(ctx, fmt.Sprintf("response status: %d", resp.StatusCode))
	reportLogger.Debug(ctx, fmt.Sprintf("response body: %s", string(body)))

	var writeFile ResponseJSON
	err = json.Unmarshal(body, &writeFile)
	if err != nil {
		reportLogger.Error(ctx, "failed parsing body: "+err.Error())
		return report.NewApiReportErr(r), err
	}

	var reportState report.State
	if resp.StatusCode == http.StatusOK && writeFile.Success {
		reportState = report.Pass
	} else if resp.StatusCode == http.StatusBadRequest {
		reportState = report.Inconclusive
	} else {
		reportState = report.Fail
	}

	apiReport := report.ApiReport{
		Name:         r.Name,
		LatencyMS:    later.Sub(now).Nanoseconds() / int64(time.Millisecond),
		ReportState:  reportState,
		Report:       reportLogger.Collect(),
		CreatedAtSec: report.NowUTCMinute().Unix(),
	}

	reportLogger.Info(ctx, "ran", r.Name)
	return apiReport, err
}
