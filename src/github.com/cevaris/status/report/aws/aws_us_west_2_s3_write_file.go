package aws

import (
	"context"
	"fmt"
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

// AwsUsWest2S3WriteFile writes new file to S3
// https://gist.github.com/CarterTsai/47f732121b34399d13fbd5765b3e11ed
func AwsUsWest2S3WriteFile(ctx context.Context, r report.Request) (report.ApiReport, error) {
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

	msg := fmt.Sprintf("some data %d", now.Unix())
	tmpFile, err := report.CreateTmpFile(msg)
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
	uploader := s3manager.NewUploaderWithClient(svc)

	f, err := os.Open(tmpFile.Name())
	if err != nil {
		reportLogger.Error(ctx, fmt.Sprintf("failed to open file %q, %v", tmpFile.Name(), err))
		return report.NewApiReportErr(r), err
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:  aws.String("status.api.report"),
		Key:     aws.String(fmt.Sprintf("aws_us_west_2_s3/write_file_%d.txt", now.Unix())),
		Body:    f,
		Expires: aws.Time(now.Add(30 * 24 * time.Hour)), // expire 30 days from now
	})
	if err != nil {
		reportLogger.Error(ctx, fmt.Sprintf("failed to upload file, %v", err))
		return report.NewApiReportErr(r), err
	}

	reportLogger.Info(ctx, fmt.Sprintf("file uploaded to, %s", result.Location))

	apiReport := report.ApiReport{
		Name:        r.Name,
		Latency:     time.Since(now),
		ReportState: report.Pass,
		Report:      reportLogger.Collect(),
		CreatedAt:   r.TimeMinute,
	}

	time.Now().UnixNano()
	reportLogger.Info(ctx, "ran", r.Name)
	return apiReport, nil
}
