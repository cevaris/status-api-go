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

const (
	// Timeout each report has to run
	RunnerTotalTimeout = 50 * time.Second
)

var projectID string
var dsClient *datastore.Client

var runnerLoggger = logging.FileLogger("runner")

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	projectID = os.Getenv("PROJECT_ID")

	_, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	runnerLoggger.Info(ctx, "starting runner...")
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
func launchRunner(ctx context.Context, r *report.Request, fn func(context.Context, *report.Request) (report.ApiReport, error)) (apiReport report.ApiReport, err error) {
	chApiReport := make(chan ChApiReport, 1)

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				runnerLoggger.Error(ctx, "Recovered in f", r.Name, rec)
				chApiReport <- ChApiReport{
					apiReport: report.NewApiReportErr(r.Name, r.ReportLogger),
					err:       errors.New(fmt.Sprintf("panic thrown in %s", r.Name)),
				}
			}
		}()

		apiReport, err := fn(ctx, r)
		chApiReport <- ChApiReport{apiReport, err}
	}()

	select {
	case chApiReport := <-chApiReport:
		return chApiReport.apiReport, chApiReport.err
	case <-ctx.Done():
		runnerLoggger.Info(ctx, r.Name, "timed out")

		return report.NewApiReportErr(r.Name, r.ReportLogger), ctx.Err()
	}
}

func launchScheduler(ctx context.Context, wg *sync.WaitGroup, name string, reportNumber int) {
	defer wg.Done()
	runnerLoggger.Info(ctx, "initial runner delay", name)
	delay(time.Second * time.Duration(reportNumber%60))
	runnerLoggger.Info(ctx, "loading runner", name)

	logger := logging.FileLogger(name)

	duration := time.Duration(60 * time.Second)
	for ; true; <-time.Tick(duration) {
		ctx, cancel := context.WithTimeout(context.Background(), RunnerTotalTimeout)

		request := report.NewRequest(logger, name)
		reportLogger := request.ReportLogger

		reportLogger.Info(ctx, "launching runner", name)

		apiReport, err := launchRunner(ctx, request, status.APIReportCatalog[name])
		if err != nil {
			if err == context.DeadlineExceeded {
				reportLogger.Error(ctx, "FAILED TIMEOUT report", name, err, apiReport)
			} else {
				reportLogger.Error(ctx, "FAILED report", name, err, apiReport)
			}
		} else {
			reportLogger.Info(ctx, "SUCCESS report", name, apiReport)
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

	runnerLoggger.Info(ctx, "started runners")

	// block so we do not exit
	// we dont expect the routines to complete
	wg.Wait()
}
