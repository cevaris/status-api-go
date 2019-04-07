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
	"github.com/cevaris/timber"
)

// {"success":true,"key":"tt67yI","link":"https://file.io/tt67yI","expiry":"14 days"}
type witeTextResponse struct {
	Success bool `json:"success"`
}

// WriteTextReport reports on writing a message to https://www.file.io
func WriteTextReport() {
	logger := timber.NewOpLogger("fileio_write_text")
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
	}
	defer resp.Body.Close()

	later := time.Now().UTC()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, err)
	}
	reportLog = append(reportLog, fmt.Sprintf("response status: %d", resp.StatusCode))
	reportLog = append(reportLog, fmt.Sprintf("response body: %s", body))

	var writeTextRes witeTextResponse
	err = json.Unmarshal(body, &writeTextRes)
	if err != nil {
		reportLog = append(reportLog, "failed parsing body: "+err.Error())
		logger.Error(ctx, err)
	}

	var reportState report.ReportState
	if resp.StatusCode == http.StatusOK && writeTextRes.Success {
		reportState = report.Pass
	} else if resp.StatusCode == http.StatusBadRequest {
		reportState = report.Inconclusive
	} else {
		reportState = report.Fail
	}

	testReport := report.ApiTestReport{
		LatencyMS:    later.Sub(now).Nanoseconds() / int64(time.Millisecond),
		ReportState:  reportState,
		Report:       strings.Join(reportLog[:], "\n"),
		CreatedAtSec: now.Unix(),
	}

	logger.Info(ctx, "ran WriteTextReport\n", testReport)
}
