package status

import (
	"github.com/cevaris/status/report"
	"github.com/cevaris/status/report/aws"
	"github.com/cevaris/status/report/fail"
	"github.com/cevaris/status/report/fileio"
)

// APIReportCatalog Report catalog
var APIReportCatalog = map[string]func(string) (report.ApiReport, error){
	"aws_us_west_2_s3_write_file": aws.AwsUsWest2S3WriteFile,
	"aws_us_west_2_s3_read_file": aws.AwsUsWest2S3ReadFile,
	"fail_http": fail.HTTPErrorReport,
	"fail_panic": fail.PanicReport,
	"fail_timeout": fail.TimeoutErrorReport,
	"fileio_write_text": fileio.WriteTextReport,
	"fileio_write_file": fileio.WriteFileReport,
}
