package pkg

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Utils interface {
	CreateUserBucket(userID string) (string, error)
}

// S3Uploader interface for manager upload
type S3Uploader interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

// S3Downloader interface for manager download
type S3Downloader interface {
	Download(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*manager.Downloader)) (n int64, err error)
}
