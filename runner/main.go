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

type ChApiReport struct {
	apiReport report.ApiReport
	err       error
}

// launchRunner executes the report runner while handling timeouts and panics
func launchRunner(ctx context.Context, name string, fn func(context.Context, string) (report.ApiReport, error)) (apiReport report.ApiReport, err error) {
	chApiReport := make(chan ChApiReport, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(ctx, "Recovered in f", name, r)
				chApiReport <- ChApiReport{
					apiReport: report.NewError(name, report.NewLogger(logger)),
					err:       errors.New(fmt.Sprintf("panic thrown in %s", name)),
				}
			}
		}()

		apiReport, err := fn(ctx, name)
		chApiReport <- ChApiReport{apiReport, err}
	}()

	select {
	case chApiReport := <-chApiReport:
		return chApiReport.apiReport, chApiReport.err
	case <-ctx.Done():
		logger.Info(ctx, name, "timed out")

		return report.NewError(name, report.NewLogger(logger)), ctx.Err()
	}
}

func launchScheduler(ctx context.Context, wg *sync.WaitGroup, name string, reportNumber int) {
	defer wg.Done()
	logger.Info(ctx, "initial runner delay", name)
	delay(time.Second * time.Duration(reportNumber%60))
	logger.Info(ctx, "loading runner", name)

	duration := time.Duration(60 * time.Second)
	for ; true; <-time.Tick(duration) {
		ctx, cancel := context.WithTimeout(context.Background(), report.RunnerTotalTimeout)
		logger.Info(ctx, "launching runner", name)

		apiReport, err := launchRunner(ctx, name, status.APIReportCatalog[name])
		if err != nil {
			if err == context.DeadlineExceeded {
				logger.Error(ctx, "FAILED TIMEOUT report", name, err, apiReport)
			} else {
				logger.Error(ctx, "FAILED report", name, err, apiReport)
			}
		} else {
			logger.Info(ctx, "SUCCESS report", name, apiReport)
		}

		// manually defer is fine, as reports "should" always finish executing
		cancel()
	}

	// we want to know when report runners fail
	panic("runner died :( " + name)
}

// https://play.golang.org/p/u2s7gNZvMOG
func launch(ctx context.Context) {
	var wg sync.WaitGroup
	var curr = 0
	for runnerName := range status.APIReportCatalog {
		wg.Add(1)

		go launchScheduler(ctx, &wg, runnerName, curr)

		curr = curr + 1
	}

	logger.Info(ctx, "started runners")

	// block so we do not exit
	// we dont expect the routines to complete
	wg.Wait()
}
