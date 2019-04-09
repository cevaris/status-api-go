package fileio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cevaris/status/logging"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cevaris/status/report"
)

// WriteFileReport reports on writing a message to https://www.file.io
func WriteFileReport(name string) (report.ApiReport, error) {
	logger := logging.Logger()
	reportLogger := report.NewLogger(logger)
	ctx := context.Background()
	now := time.Now().UTC()

	reportLogger.Debug(ctx, "starting test")

	msg := fmt.Sprintf("secret number %d", now.Unix())
	tmpFile, err := createTmpFile(msg)
	if err != nil {
		reportLogger.Error(ctx, "failed creating temp file: "+err.Error())
		return report.NewError(name, reportLogger), err
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			reportLogger.Error(ctx, "failed to remove temp file", err)
		}
	}()

	resp, err := uploadFile(ctx, "https://file.io", tmpFile.Name())
	if err != nil {
		reportLogger.Debug(ctx, "starting failed: "+err.Error())
		reportLogger.Error(ctx, err)
		return report.NewError(name, reportLogger), err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			reportLogger.Error(ctx, "failed to remove temp file", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reportLogger.Error(ctx, err)
		return report.NewError(name, reportLogger), err
	}
	reportLogger.Debug(ctx, fmt.Sprintf("response status: %d", resp.StatusCode))
	reportLogger.Debug(ctx, fmt.Sprintf("response body: %s", string(body)))

	var writeText ResponseJSON
	err = json.Unmarshal(body, &writeText)
	if err != nil {
		reportLogger.Debug(ctx, "failed parsing body: "+err.Error())
		return report.NewError(name, reportLogger), err
	}

	var reportState report.ReportState
	if resp.StatusCode == http.StatusOK && writeText.Success {
		reportState = report.Pass
	} else if resp.StatusCode == http.StatusBadRequest {
		reportState = report.Inconclusive
	} else {
		reportState = report.Fail
	}

	later := time.Now().UTC()
	apiReport := report.ApiReport{
		Name:         name,
		LatencyMS:    later.Sub(now).Nanoseconds() / int64(time.Millisecond),
		ReportState:  reportState,
		Report:       strings.Join(reportLogger.Collect()[:], "\n"),
		CreatedAtSec: report.NowUTCMinute().Unix(),
	}

	reportLogger.Info(ctx, "ran", name)
	return apiReport, nil
}

func uploadFile(ctx context.Context, postURL string, filename string) (*http.Response, error) {
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
	defer func() {
		if err := fh.Close(); err != nil {
			fmt.Println(ctx, "defer: failed to open file", filename, err)
		}
	}()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return nil, err
	}

	contentType := bodyWriter.FormDataContentType()
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}

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
