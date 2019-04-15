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

var logger = logging.CachedLogger("runner", true)

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
func launchRunner(ctx context.Context, r report.Request, fn func(context.Context, report.Request) (report.ApiReport, error)) (apiReport report.ApiReport, err error) {
	chApiReport := make(chan ChApiReport, 1)

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Info(ctx, "Recovered in f", r.Name, rec)
				chApiReport <- ChApiReport{
					apiReport: report.NewApiReportErr(r),
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
		logger.Info(ctx, r.Name, "timed out")
		return report.NewApiReportErr(r), ctx.Err()
	}
}

func launchScheduler(ctx context.Context, wg *sync.WaitGroup, reportName string, reportNumber int) {
	defer wg.Done()
	logger.Info(ctx, "initial runner delay", reportName)
	// mod 55 as we dont want jobs running at the last 5 seconds,
	// as it could overwrite the future minute's report if the job finishes near and after 59th second.
	delay(time.Second * time.Duration(reportNumber%55))
	logger.Info(ctx, "loading runner", reportName)

	localLogger := logging.CachedLogger(reportName, false)

	duration := time.Duration(60 * time.Second)
	for ; true; <-time.Tick(duration) {
		ctx, cancel := context.WithTimeout(context.Background(), RunnerTotalTimeout)

		request := report.NewRequest(localLogger, reportName)
		reportLogger := request.ReportLogger

		reportLogger.Info(ctx, "launching runner", reportName)

		apiReport, err := launchRunner(ctx, request, status.APIReportCatalog[reportName])
		writeReport(ctx, request, apiReport, err)

		// manually defer is fine, as reports "should" always finish executing
		cancel()
	}

	// we want to know when report runners fail
	panic("runner died :( " + reportName)
}

func writeReport(ctx context.Context, r report.Request, apiReport report.ApiReport, err error) {
	reportLogger := r.ReportLogger

	reportLogger.Info(ctx, "report.createdAt.sec", apiReport.CreatedAt)
	reportLogger.Info(ctx, "report.latency.ms", apiReport.Latency)
	reportLogger.Info(ctx, "report.name", apiReport.Name)

	if err != nil {
		if err == context.DeadlineExceeded {
			reportLogger.Error(ctx, "report.state", apiReport.ReportState, "TIMEOUT")
		} else {
			reportLogger.Error(ctx, "report.state", apiReport.ReportState, err)
		}
	} else {
		reportLogger.Error(ctx, "report.state", apiReport.ReportState)
	}

	//logger.Info(ctx, "report.log", len(reportLogger.Collect()), "bytes")
}

// https://play.golang.org/p/u2s7gNZvMOG
func launch(ctx context.Context) {
	var wg sync.WaitGroup
	var curr = 0
	for reportName := range status.APIReportCatalog {
		wg.Add(1)

		go launchScheduler(ctx, &wg, reportName, curr)

		curr = curr + 1
	}

	logger.Info(ctx, "started runners")

	// block so we do not exit
	// we dont expect the routines to complete
	wg.Wait()
}
