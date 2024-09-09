package mysql

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type SeedDataStore struct {
	BucketName string `json:"bucket_name"`
	Path       string `json:"path"`
	FileName   string `json:"file_name"`
}

func Seeder(ctx context.Context, cfg DBConfig, awsCfg aws.Config, dataStore SeedDataStore) error {
	return nil
}
