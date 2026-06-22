package attachments

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const maxUploadSizeBytes = 10 * 1024 * 1024 // 10 MB

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

func buildObjectKey(feedbackID, attachmentID, fileName string) string {
	cleanFileName := sanitizeFileName(fileName)

	return fmt.Sprintf(
		"feedback/%s/attachments/%s/%s",
		feedbackID,
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
