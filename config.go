package status

import (
	"github.com/cevaris/status/report"
	"github.com/cevaris/status/report/fileio"
)

var ApiTestStore = map[string]int64{
	"fileio_write_text": 0,
	"fileio_write_file": 0,
}

var Lookup = map[string]func(string) (report.ApiReport, error){
	"fileio_write_text": fileio.WriteTextReport,
	"fileio_write_file": fileio.WriteFileReport,
}
