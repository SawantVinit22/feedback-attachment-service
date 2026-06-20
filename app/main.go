package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	maxUploadSizeBytes = 10 * 1024 * 1024 // 10 MB
	presignExpiry      = 10 * time.Minute
)

type PresignUploadRequest struct {
	FeedbackID  string `json:"feedback_id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type PresignUploadResponse struct {
	UploadURL        string            `json:"upload_url"`
	ObjectKey        string            `json:"object_key"`
	ExpiresInSeconds int64             `json:"expires_in_seconds"`
	RequiredHeaders  map[string]string `json:"required_headers"`
}

type S3PresignService struct {
	bucket    string
	presigner *s3.PresignClient
}

func NewS3PresignService(ctx context.Context, bucket string) (*S3PresignService, error) {
	if strings.TrimSpace(bucket) == "" {
		return nil, errors.New("S3_BUCKET_NAME is required")
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &S3PresignService{
		bucket:    bucket,
		presigner: s3.NewPresignClient(s3Client),
	}, nil
}

func (s *S3PresignService) GenerateUploadURL(
	ctx context.Context,
	req PresignUploadRequest,
	userID string,
) (*PresignUploadResponse, error) {
	if err := validatePresignRequest(req); err != nil {
		return nil, err
	}

	objectKey := buildObjectKey(req.FeedbackID, userID, req.FileName)

	presignedReq, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(req.ContentType),
		Metadata: map[string]string{
			"original-file-name": req.FileName,
			"feedback-id":        req.FeedbackID,
			"uploaded-by":        userID,
		},
	}, func(opts *s3.PresignOptions) {
		opts.Expires = presignExpiry
	})

	if err != nil {
		return nil, fmt.Errorf("generate presigned put url: %w", err)
	}

	return &PresignUploadResponse{
		UploadURL:        presignedReq.URL,
		ObjectKey:        objectKey,
		ExpiresInSeconds: int64(presignExpiry.Seconds()),
		RequiredHeaders: map[string]string{
			"Content-Type": req.ContentType,
		},
	}, nil
}

func validatePresignRequest(req PresignUploadRequest) error {
	if strings.TrimSpace(req.FeedbackID) == "" {
		return errors.New("feedback_id is required")
	}

	if strings.TrimSpace(req.FileName) == "" {
		return errors.New("file_name is required")
	}

	if strings.TrimSpace(req.ContentType) == "" {
		return errors.New("content_type is required")
	}

	if req.SizeBytes <= 0 {
		return errors.New("size_bytes must be greater than zero")
	}

	if req.SizeBytes > maxUploadSizeBytes {
		return fmt.Errorf("file size exceeds max limit of %d bytes", maxUploadSizeBytes)
	}

	if !isAllowedContentType(req.ContentType) {
		return fmt.Errorf("content_type %s is not allowed", req.ContentType)
	}

	return nil
}

func isAllowedContentType(contentType string) bool {
	allowed := map[string]bool{
		"image/png":       true,
		"image/jpeg":      true,
		"application/pdf": true,
		"text/plain":      true,
	}

	return allowed[strings.ToLower(contentType)]
}

func buildObjectKey(feedbackID, userID, fileName string) string {
	cleanFileName := sanitizeFileName(fileName)
	attachmentID := uuid.NewString()

	return fmt.Sprintf(
		"feedback/%s/users/%s/%s-%s",
		feedbackID,
		userID,
		attachmentID,
		cleanFileName,
	)
}

func sanitizeFileName(fileName string) string {
	base := filepath.Base(fileName)
	base = strings.ToLower(base)

	re := regexp.MustCompile(`[^a-z0-9._-]+`)
	base = re.ReplaceAllString(base, "_")

	return base
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write json response: %v", err)
	}
}

func presignUploadHandler(service *S3PresignService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "method not allowed",
			})
			return
		}

		var req PresignUploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid json request body",
			})
			return
		}

		// Temporary hardcoded user.
		// Later this will come from JWT/Cognito authentication middleware.
		userID := "user_123"

		resp, err := service.GenerateUploadURL(r.Context(), req, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func main() {
	ctx := context.Background()

	bucket := os.Getenv("S3_BUCKET_NAME")

	service, err := NewS3PresignService(ctx, bucket)
	if err != nil {
		log.Fatalf("failed to create s3 presign service: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/attachments/presign-upload", presignUploadHandler(service))

	addr := ":8080"
	log.Printf("server started on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
