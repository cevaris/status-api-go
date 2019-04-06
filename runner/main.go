package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

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

		fmt.Println("work!")

		sleepDuration := nextSleepDuration()
		fmt.Println("sleeping for ", sleepDuration, "seconds")
		time.Sleep(sleepDuration)
	}
}
