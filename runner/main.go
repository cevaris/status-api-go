package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
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
					f(k) // launch test
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
