package main

import (
	"context"
	"errors"
	"fmt"
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

func ignorePanic(ctx context.Context, name string, fn func(string) (report.ApiReport, error)) (apiReport report.ApiReport, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "Recovered in f", name, r)
			apiReport = report.NewError(name, report.NewLogger(logger))
			err = errors.New(fmt.Sprintf("panic thrown in %s", name))
			return
		}
	}()

	return fn(name)
}

func periodicReport(name string, duration time.Duration, fn func(string) (report.ApiReport, error)) {
	for ; true; <-time.Tick(duration) {
		ctx := context.Background()
		logger.Info(ctx, "launching runner", name)
		apiReport, err := ignorePanic(ctx, name, fn)
		if err != nil {
			logger.Error(ctx, "FAILED report", name, err, apiReport)
		} else {
			logger.Info(ctx, "SUCCESS report", name, apiReport)
		}
	}

	// we want to know when report runners fail
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
