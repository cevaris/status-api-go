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
	"time"

	"github.com/cevaris/status/report"
)

// WriteFileReport reports on writing a message to https://www.file.io
func WriteFileReport(ctx context.Context, r report.Request) (report.ApiReport, error) {
	reportLogger := r.ReportLogger
	now := time.Now().UTC()

	reportLogger.Debug(ctx, "starting test")

	msg := fmt.Sprintf("secret number %d", now.Unix())
	tmpFile, err := report.CreateTmpFile(msg)
	if err != nil {
		reportLogger.Error(ctx, "failed creating temp file: "+err.Error())
		return report.NewApiReportErr(r.Name, reportLogger), err
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
		return report.NewApiReportErr(r.Name, reportLogger), err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			reportLogger.Error(ctx, "failed to remove temp file", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reportLogger.Error(ctx, err)
		return report.NewApiReportErr(r.Name, reportLogger), err
	}
	reportLogger.Debug(ctx, fmt.Sprintf("response status: %d", resp.StatusCode))
	reportLogger.Debug(ctx, fmt.Sprintf("response body: %s", string(body)))

	var writeText ResponseJSON
	err = json.Unmarshal(body, &writeText)
	if err != nil {
		reportLogger.Debug(ctx, "failed parsing body: "+err.Error())
		return report.NewApiReportErr(r.Name, reportLogger), err
	}

	var reportState report.State
	if resp.StatusCode == http.StatusOK && writeText.Success {
		reportState = report.Pass
	} else if resp.StatusCode == http.StatusBadRequest {
		reportState = report.Inconclusive
	} else {
		reportState = report.Fail
	}

	later := time.Now().UTC()
	apiReport := report.ApiReport{
		Name:         r.Name,
		LatencyMS:    later.Sub(now).Nanoseconds() / int64(time.Millisecond),
		ReportState:  reportState,
		Report:       reportLogger.Collect(),
		CreatedAtSec: report.NowUTCMinute().Unix(),
	}

	reportLogger.Info(ctx, "ran", r.Name)
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
