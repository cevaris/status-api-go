package fileio

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cevaris/status/report"
)

// WriteTextReport reports on writing a message to https://www.file.io
func WriteTextReport(name string) (report.ApiReport, error) {
	var apiReport report.ApiReport

	ctx := context.Background()
	now := time.Now().UTC()

	reportLog := make([]string, 0)
	reportLog = append(reportLog, "starting test")

	data := url.Values{}
	data.Add("text", fmt.Sprintf("secret number %d", now.Unix()))

	resp, err := http.PostForm("https://file.io", data)
	if err != nil {
		reportLog = append(reportLog, "starting failed: "+err.Error())
		logger.Error(ctx, err)
		return apiReport, err
	}
	defer resp.Body.Close()

	later := time.Now().UTC()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, err)
		return apiReport, err
	}
	reportLog = append(reportLog, fmt.Sprintf("response status: %d", resp.StatusCode))
	reportLog = append(reportLog, fmt.Sprintf("response body: %s", body))

	var writeFile ResponseJSON
	err = json.Unmarshal(body, &writeFile)
	if err != nil {
		reportLog = append(reportLog, "failed parsing body: "+err.Error())
		logger.Error(ctx, err)
		return apiReport, err
	}

	var reportState report.ReportState
	if resp.StatusCode == http.StatusOK && writeFile.Success {
		reportState = report.Pass
	} else if resp.StatusCode == http.StatusBadRequest {
		reportState = report.Inconclusive
	} else {
		reportState = report.Fail
	}

	apiReport = report.ApiReport{
		LatencyMS:    later.Sub(now).Nanoseconds() / int64(time.Millisecond),
		ReportState:  reportState,
		Report:       strings.Join(reportLog[:], "\n"),
		CreatedAtSec: now.Unix(),
	}

	logger.Info(ctx, "ran", name, fmt.Sprintf("%+v", apiReport))
	return apiReport, err
}
