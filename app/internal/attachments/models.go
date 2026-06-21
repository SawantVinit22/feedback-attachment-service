package attachments

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

type CompleteUploadRequest struct {
	FeedbackID string `json:"feedback_id"`
	ObjectKey  string `json:"object_key"`
}

type CompleteUploadResponse struct {
	ObjectKey   string            `json:"object_key"`
	Status      string            `json:"status"`
	SizeBytes   int64             `json:"size_bytes"`
	ContentType string            `json:"content_type"`
	Metadata    map[string]string `json:"metadata"`
}

type PresignDownloadRequest struct {
	FeedbackID string `json:"feedback_id"`
	ObjectKey  string `json:"object_key"`
}

type PresignDownloadResponse struct {
	DownloadURL      string `json:"download_url"`
	ObjectKey        string `json:"object_key"`
	ExpiresInSeconds int64  `json:"expires_in_seconds"`
	SizeBytes        int64  `json:"size_bytes"`
	ContentType      string `json:"content_type"`
}
