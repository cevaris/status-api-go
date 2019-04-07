package fileio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cevaris/status/report"
)

// {"success":true,"key":"tt67yI","link":"https://file.io/tt67yI","expiry":"14 days"}
type witeTextResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// WriteFileReport reports on writing a message to https://www.file.io
func WriteFileReport(name string) (report.ApiReport, error) {
	var apiReport report.ApiReport
	ctx := context.Background()
	now := time.Now().UTC()

	reportLog := make([]string, 0)
	reportLog = append(reportLog, "starting test")

	msg := fmt.Sprintf("secret number %d", now.Unix())
	tmpFile, err := createTmpFile(msg)
	if err != nil {
		reportLog = append(reportLog, "failed creating temp file: "+err.Error())
		logger.Error(ctx, err)
		return apiReport, err
	}
	defer os.Remove(tmpFile.Name())

	resp, err := uploadFile("https://file.io", tmpFile.Name())
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

	var writeText ResponseJSON
	err = json.Unmarshal(body, &writeText)
	if err != nil {
		reportLog = append(reportLog, "failed parsing body: "+err.Error())
		logger.Error(ctx, err)
		return apiReport, err
	}

	var reportState report.ReportState
	if resp.StatusCode == http.StatusOK && writeText.Success {
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
	return apiReport, nil
}

func uploadFile(postURL string, filename string) (*http.Response, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return nil, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	return http.Post(postURL, contentType, bodyBuf)
}

func createTmpFile(msg string) (*os.File, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "runner-")
	if err != nil {
		return nil, err
	}

	// Example writing to the file
	text := []byte(msg)
	if _, err = tmpFile.Write(text); err != nil {
		return nil, err
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return tmpFile, nil
}
