package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cevaris/status"
	"github.com/cevaris/timber"

	"cloud.google.com/go/datastore"
)

var projectID string
var dsClient *datastore.Client
var logger timber.Logger

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	projectID = os.Getenv("PROJECT_ID")

	logger = timber.NewOpLogger("runner")

	_, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	fmt.Println("starting runner...")
	go forever()
	select {} // block forever
}

// nextWaitSecs should check at least once a minute for new jobs to execute
func nextSleepDuration() time.Duration {
	return time.Second * (25 + time.Duration(rand.Intn(5)))
}

func forever() {
	fmt.Println("started runner")
	for {
		// do work

		for k, lastRanSec := range status.ApiTestStore {
			now := time.Now().Unix()
			if now-int64(lastRanSec) > 60 {
				// time to run again
				if f, ok := status.Lookup[k]; ok {
					f() // launch test
					status.ApiTestStore[k] = time.Now().Unix()
				}
			}
		}

		ctx := context.Background()
		logger.Info(ctx, "work")

		sleepDuration := nextSleepDuration()
		fmt.Println("sleeping for ", sleepDuration, "seconds")
		time.Sleep(sleepDuration)
	}
}

func formatRequest(r *http.Request) string {
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
