package report

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

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

func CreateEmptyTmpFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "runner-")
}

func CreateTmpFile(msg string) (*os.File, error) {
	tmpFile, err := CreateEmptyTmpFile()
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

// NowMinute returns now, truncated down to the minute. Useful for timestamping with minute grainularity.
// https://play.golang.org/p/cpW3itpYHia
func NowUTCMinute() time.Time {
	now := time.Now().UTC()
	return UTCMinute(now)
}

func UTCMinute(t time.Time) time.Time {
	return t.UTC().Truncate(60 * time.Second)
}
