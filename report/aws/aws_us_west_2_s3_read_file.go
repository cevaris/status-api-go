package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cevaris/status/logging"
	"github.com/cevaris/status/report"
	"github.com/cevaris/status/secrets"
	"os"
	"time"
)


// AwsUsWest2S3ReadFile downloads a existing file from S3
func AwsUsWest2S3ReadFile(name string) (report.ApiReport, error) {
	logger := logging.Logger()
	reportLogger := report.NewLogger(logger)
	ctx := context.Background()
	now := time.Now().UTC()

	apiKeys := secrets.ReadOnlyApiKeys
	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     apiKeys.AwsAccessKeyID,
			SecretAccessKey: apiKeys.AwsSecretAccessKey,
		}),
		Region: aws.String("us-west-2"),
	})
	
	tmpFile, err := report.CreateEmptyTmpFile()
	if err != nil {
		reportLogger.Error(ctx, "failed creating temp file: "+err.Error())
		return report.NewError(name, reportLogger), err
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			reportLogger.Error(ctx, "failed to remove temp file", err)
		}
	}()

	// Create an uploader with the session and default options
	downloader := s3manager.NewDownloaderWithClient(svc)

	// Upload the file to S3.
	bytes, err := downloader.Download(tmpFile, &s3.GetObjectInput{
		Bucket: aws.String("status.api.report"),
		Key:    aws.String("aws_us_west_2_s3/readonly.txt"),
	})
	if err != nil {
		reportLogger.Error(ctx, "failed to read file", err)
		return report.NewError(name, reportLogger), err
	}

	reportLogger.Info(ctx, "file uploaded to ", tmpFile.Name(), "downloaded", bytes, "bytes")

	later := time.Now().UTC()
	apiReport := report.ApiReport{
		Name:         name,
		LatencyMS:    later.Sub(now).Nanoseconds() / int64(time.Millisecond),
		ReportState:  report.Pass,
		Report:       reportLogger.Collect(),
		CreatedAtSec: report.NowUTCMinute().Unix(),
	}

	reportLogger.Info(ctx, "ran", name)
	return apiReport, nil
}
