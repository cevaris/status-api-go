package status

import (
	"github.com/cevaris/status/report"
	"github.com/cevaris/status/report/aws"
)

// APIReportCatalog Report catalog
var APIReportCatalog = map[string]func(string) (report.ApiReport, error){
	"aws_us_west_2_s3_write_file": aws.AwsUsWest2S3WriteFile,
	//"fail_http": fail.HTTPErrorReport,
	//"fail_panic": fail.PanicReport,
	//"fileio_write_text": fileio.WriteTextReport,
	//"fileio_write_file": fileio.WriteFileReport,
}
