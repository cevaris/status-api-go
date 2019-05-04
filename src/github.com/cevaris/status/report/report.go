package report

import (
	"time"
)

// State state of the report
type State int

const (
	// States of which a report can result to
	Pass         State = 0
	Inconclusive State = 1
	Fail         State = 2
)

// ApiReport is used for code/logic in-memory usage
type ApiReport struct {
	CreatedAt   time.Time
	Latency     time.Duration
	Name        string
	Report      []byte
	ReportState State
}

// ApiReportJson is for rendering out JSON over the API
type ApiReportJson struct {
	CreatedAt   string `json:"created_at"`
	Latency     string `json:"latency"`
	Name        string `json:"name"`
	ReportState string `json:"report_state"`
}

// ApiReportRecord is written to disk
type ApiReportRecord struct {
	CreatedAt   time.Time
	Latency     time.Duration
	Name        string
	Report      []byte `datastore:",noindex"`
	ReportState State
}

func (r *ApiReport) Present() ApiReportJson {
	return ApiReportJson{
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		Latency:     r.Latency.String(),
		Name:        r.Name,
		ReportState: r.ReportState.Name(),
	}
}

func (r *ApiReport) Record() ApiReportRecord {
	return ApiReportRecord{
		CreatedAt:   r.CreatedAt,
		Latency:     r.Latency,
		Name:        r.Name,
		Report:      r.Report,
		ReportState: r.ReportState,
	}
}

func (r *ApiReportRecord) Lift() ApiReport {
	return ApiReport{
		CreatedAt:   r.CreatedAt,
		Latency:     r.Latency,
		Name:        r.Name,
		Report:      r.Report,
		ReportState: r.ReportState,
	}
}

func (s State) Name() string {
	switch s {
	case Pass:
		return "PASS"
	case Inconclusive:
		return "INCONCLUSIVE"
	case Fail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

func NewApiReportErr(r Request) ApiReport {
	return ApiReport{
		Name:        r.Name,
		Latency:     0,
		ReportState: Fail,
		Report:      r.ReportLogger.Collect(),
		CreatedAt:   NowUTCMinute(),
	}
}
