package report

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// {"success":true,"key":"tt67yI","link":"https://file.io/tt67yI","expiry":"14 days"}
type witeTextResponse struct {
	Success bool `json:"success"`
}

// FileioWriteText reports on writing a message to https://www.file.io
func FileioWriteText() {
	ctx := context.Background()
	now := time.Now().UTC()

	report := make([]string, 0)
	report = append(report, "starting test")

	data := url.Values{}
	data.Add("text", fmt.Sprintf("secret number %d", now.Unix()))

	resp, err := http.PostForm("https://file.io", data)
	if err != nil {
		report = append(report, "starting failed: "+err.Error())
		logger.Error(ctx, err)
	}
	defer resp.Body.Close()

	later := time.Now().UTC()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, err)
	}
	report = append(report, fmt.Sprintf("response status: %d", resp.StatusCode))
	report = append(report, fmt.Sprintf("response body: %s", body))

	var writeTextRes witeTextResponse
	err = json.Unmarshal(body, &writeTextRes)
	if err != nil {
		report = append(report, "failed parsing body: "+err.Error())
		logger.Error(ctx, err)
	}

	var testState TestResultState
	if resp.StatusCode == http.StatusOK && writeTextRes.Success {
		testState = Pass
	} else if resp.StatusCode == http.StatusBadRequest {
		testState = Inconclusive
	} else {
		testState = Fail
	}

	testReport := ApiTestReport{
		LatencyMS:    later.Sub(now).Nanoseconds() / int64(time.Millisecond),
		TestState:    testState,
		Report:       strings.Join(report[:], "\n"),
		CreatedAtSec: now.Unix(),
	}

	logger.Info(ctx, "ran fileioWriteText\n", testReport)
}
