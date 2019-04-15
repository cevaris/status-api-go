package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cevaris/status/report"
	"github.com/cevaris/status/secrets"
	"os"
	"time"
)

// AwsUsWest2S3ReadFile downloads a existing file from S3
func AwsUsWest2S3ReadFile(ctx context.Context, r report.Request) (report.ApiReport, error) {
	reportLogger := r.ReportLogger
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
		return report.NewApiReportErr(r), err
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
		return report.NewApiReportErr(r), err
	}

	reportLogger.Info(ctx, "file uploaded to ", tmpFile.Name(), "downloaded", bytes, "bytes")

	apiReport := report.ApiReport{
		Name:        r.Name,
		Latency:     time.Since(now),
		ReportState: report.Pass,
		Report:      reportLogger.Collect(),
		CreatedAt:   r.TimeMinute,
	}

	reportLogger.Info(ctx, "ran", r.Name)
	return apiReport, nil
}
