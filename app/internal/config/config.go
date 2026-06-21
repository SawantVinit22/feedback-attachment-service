package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	S3BucketName string
	ServerAddr   string
}

func Load() (*Config, error) {
	s3BucketName := strings.TrimSpace(os.Getenv("S3_BUCKET_NAME"))
	if s3BucketName == "" {
		return nil, errors.New("S3_BUCKET_NAME is required")
	}

	serverAddr := strings.TrimSpace(os.Getenv("SERVER_ADDR"))
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	return &Config{
		S3BucketName: s3BucketName,
		ServerAddr:   serverAddr,
	}, nil
}
