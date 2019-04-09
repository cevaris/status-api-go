package main

import (
	"context"
	"github.com/cevaris/status/logging"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/cevaris/status/report"

	"cloud.google.com/go/datastore"
	"github.com/cevaris/status"
)

var projectID string
var dsClient *datastore.Client

var logger = logging.Logger()

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	projectID = os.Getenv("PROJECT_ID")

	_, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	logger.Info(ctx, "starting runner...")
	launch(ctx)
}

func delay(duration time.Duration) {
	time.Sleep(duration)
}

func periodicReport(name string, duration time.Duration, fn func(string) (report.ApiReport, error)) {
	for ; true; <-time.Tick(duration) {
		ctx := context.Background()
		logger.Info(ctx, "launching runner", name)
		apiReport, err := fn(name)
		if err != nil {
			logger.Error(ctx, "failed to run report", name, err, apiReport)
		} else {
			logger.Info(ctx, "successful report", name, apiReport)
		}
	}

	panic("runner died :( " + name)
}

func launchImpl(ctx context.Context, wg *sync.WaitGroup, runnerName string, reportNumber int) {
	defer wg.Done()
	logger.Info(ctx, "initial runner delay", runnerName)
	delay(time.Second * time.Duration(reportNumber%60))
	logger.Info(ctx, "loading runner", runnerName)
	periodicReport(runnerName, time.Duration(60*time.Second), status.APIReportCatalog[runnerName])
}

// https://play.golang.org/p/u2s7gNZvMOG
func launch(ctx context.Context) {
	var wg sync.WaitGroup
	var curr = 0
	for runnerName := range status.APIReportCatalog {
		wg.Add(1)

		go launchImpl(ctx, &wg, runnerName, curr)

		curr = curr + 1
	}

	logger.Info(ctx, "started runners")

	// block so we do not exit
	// we dont expect the routines to complete
	wg.Wait()
}
