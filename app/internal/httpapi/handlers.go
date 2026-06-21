package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/SawantVinit22/feedback-attachment-service/app/internal/attachments"
)

type Handler struct {
	service *attachments.S3PresignService
}

func NewHandler(service *attachments.S3PresignService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/attachments/presign-upload", h.presignUploadHandler)
	mux.HandleFunc("/attachments/complete-upload", h.completeUploadHandler)
	mux.HandleFunc("/attachments/presign-download", h.presignDownloadHandler)
}

func (h *Handler) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) presignUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req attachments.PresignUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json request body",
		})
		return
	}

	// Temporary hardcoded user.
	// Later this will come from JWT/Cognito authentication middleware.
	userID := "user_123"

	resp, err := h.service.GenerateUploadURL(r.Context(), req, userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) completeUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req attachments.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json request body",
		})
		return
	}

	// Temporary hardcoded user.
	// Later this will come from JWT/Cognito authentication middleware.
	userID := "user_123"

	resp, err := h.service.CompleteUpload(r.Context(), req, userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) presignDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req attachments.PresignDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json request body",
		})
		return
	}

	// Temporary hardcoded user.
	// Later this will come from JWT/Cognito authentication middleware.
	userID := "user_123"

	resp, err := h.service.GenerateDownloadURL(r.Context(), req, userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
