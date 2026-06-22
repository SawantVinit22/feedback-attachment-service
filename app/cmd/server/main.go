package main

import (
	"context"
	"log"
	"net/http"

	"github.com/SawantVinit22/feedback-attachment-service/app/internal/attachments"
	"github.com/SawantVinit22/feedback-attachment-service/app/internal/auth"
	"github.com/SawantVinit22/feedback-attachment-service/app/internal/config"
	"github.com/SawantVinit22/feedback-attachment-service/app/internal/database"
	"github.com/SawantVinit22/feedback-attachment-service/app/internal/httpapi"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer dbPool.Close()

	log.Println("connected to postgres")

	attachmentRepository := attachments.NewRepository(dbPool)

	service, err := attachments.NewS3PresignService(ctx, cfg.S3BucketName, attachmentRepository)
	if err != nil {
		log.Fatalf("failed to create s3 presign service: %v", err)
	}

	mux := http.NewServeMux()

	handler := httpapi.NewHandler(service)
	handler.RegisterRoutes(mux)

	authMiddleware, err := auth.NewMiddleware(
		ctx,
		cfg.OIDCIssuerURL,
		cfg.OIDCClientID,
		cfg.OIDCUserClaim,
	)
	if err != nil {
		log.Fatalf("failed to create auth middleware: %v", err)
	}

	securedHandler := authMiddleware.RequireAuth(mux)

	log.Printf("server started on %s", cfg.ServerAddr)

	if err := http.ListenAndServe(cfg.ServerAddr, securedHandler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
