package main

import (
	"context"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/cevaris/status/report"

	"github.com/cevaris/status"
	"github.com/cevaris/timber"

	"cloud.google.com/go/datastore"
)

var projectID string
var dsClient *datastore.Client
var logger = timber.NewOpLogger("runner")

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
	defer panic("runner died :( " + name)

	for ; true; <-time.Tick(duration) {
		ctx := context.Background()
		logger.Info(ctx, "launching runner", name)
		report, err := fn(name)
		if err != nil {
			logger.Error(ctx, "failed to run report", name, err)
		} else {
			logger.Info(ctx, "successful report", name, report)
		}
	}
}

func launch(ctx context.Context) {
	var waitgroup sync.WaitGroup
	var curr = 0
	for runnerName, fn := range status.APIReportCatalog {
		waitgroup.Add(1)

		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			logger.Info(ctx, "initial runner delay", runnerName)
			delay(time.Second * time.Duration(i%60))
			logger.Info(ctx, "loading runner", runnerName)
			periodicReport(runnerName, time.Duration(60*time.Second), fn)
		}(&waitgroup, curr)

		curr = curr + 1
	}

	logger.Info(ctx, "started runners")

	// block so we do not exit
	// we dont expect the routines to complete
	waitgroup.Wait()
}
