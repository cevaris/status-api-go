package report

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cevaris/timber"
)

// Log is a logger for all report runners
var Log = timber.NewOpLogger("report-runnner")

// ReportState state of the report
type ReportState int

const (
	Pass         ReportState = 0
	Fail         ReportState = 1
	Inconclusive ReportState = 2
)

// ApiReport is written to disk
type ApiReport struct {
	LatencyMS    int64
	ReportState  ReportState
	Report       string
	CreatedAtSec int64
}

// NowMinute returns now, truncated down to the minute. Useful for timestamping with minute grainularity.
// https://play.golang.org/p/cpW3itpYHia
func NowUTCMinute() time.Time {
	now := time.Now().UTC()
	return now.Truncate(60 * time.Second)
}

//FmtHTTPRequest returns a formatted string of http request
func FmtHTTPRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}


func NewError(name string, reportLog []string) ApiReport {
	return ApiReport{
		LatencyMS:    0,
		ReportState:  Fail,
		Report:       strings.Join(reportLog[:], "\n"),
		CreatedAtSec: NowUTCMinute().Unix(),
	}
}