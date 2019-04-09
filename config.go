package status

import (
	"github.com/cevaris/status/report"
	"github.com/cevaris/status/report/fail"
	"github.com/cevaris/status/report/fileio"
)

// APIReportCatalog Report catalog
var APIReportCatalog = map[string]func(string) (report.ApiReport, error){
	"fail_http": fail.HTTPErrorReport,
	"fail_panic": fail.PanicReport,
	"fileio_write_text": fileio.WriteTextReport,
	"fileio_write_file": fileio.WriteFileReport,
}
