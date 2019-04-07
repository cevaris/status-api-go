package report

import (
	"time"

	"github.com/cevaris/timber"
)

var logger = timber.NewOpLogger("runner")

type TestResultState int

const (
	Pass         TestResultState = 0
	Fail         TestResultState = 1
	Inconclusive TestResultState = 2
)

// ApiTestReport is written to disk
type ApiTestReport struct {
	LatencyMS    int64
	TestState    TestResultState
	Report       string
	CreatedAtSec int64
}

// NowMinute returns now, truncated down to the minute. Useful for timestamping with minute grainularity.
// https://play.golang.org/p/cpW3itpYHia
func NowMinute() time.Time {
	now := time.Now().UTC()
	return now.Truncate(60 * time.Second)
}
