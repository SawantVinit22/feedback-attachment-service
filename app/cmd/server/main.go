package main

import (
	"context"
	"log"
	"net/http"

	"github.com/SawantVinit22/feedback-attachment-service/app/internal/attachments"
	"github.com/SawantVinit22/feedback-attachment-service/app/internal/config"
	"github.com/SawantVinit22/feedback-attachment-service/app/internal/httpapi"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	service, err := attachments.NewS3PresignService(ctx, cfg.S3BucketName)
	if err != nil {
		log.Fatalf("failed to create s3 presign service: %v", err)
	}

	mux := http.NewServeMux()

	handler := httpapi.NewHandler(service)
	handler.RegisterRoutes(mux)

	log.Printf("server started on %s", cfg.ServerAddr)

	if err := http.ListenAndServe(cfg.ServerAddr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
