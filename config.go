package status

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

var ApiTestStore = map[string]int64{
	"fileio_write_text": 0,
}
