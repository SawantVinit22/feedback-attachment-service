package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	S3BucketName string
	ServerAddr   string
	DatabaseURL  string
}

func Load() (*Config, error) {
	s3BucketName := strings.TrimSpace(os.Getenv("S3_BUCKET_NAME"))
	if s3BucketName == "" {
		return nil, errors.New("S3_BUCKET_NAME is required")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	serverAddr := strings.TrimSpace(os.Getenv("SERVER_ADDR"))
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	return &Config{
		S3BucketName: s3BucketName,
		ServerAddr:   serverAddr,
		DatabaseURL:  databaseURL,
	}, nil
}
