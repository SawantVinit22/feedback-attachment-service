package attachments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttachmentStatus string

const (
	AttachmentStatusPendingUpload AttachmentStatus = "PENDING_UPLOAD"
	AttachmentStatusUploaded      AttachmentStatus = "UPLOADED"
	AttachmentStatusFailed        AttachmentStatus = "FAILED"
	AttachmentStatusDeleted       AttachmentStatus = "DELETED"
)

var (
	ErrAttachmentNotUploaded      = errors.New("attachment not found or not uploaded")
	ErrAttachmentNotPendingUpload = errors.New("attachment not found or not pending upload")
)

type Attachment struct {
	ID                 uuid.UUID
	FeedbackID         string
	UserID             string
	ObjectKey          string
	OriginalFileName   string
	ContentType        string
	RequestedSizeBytes int64
	SizeBytes          *int64
	Status             AttachmentStatus
	CreatedAt          time.Time
	UploadedAt         *time.Time
	DeletedAt          *time.Time
	UpdatedAt          time.Time
}

type CreatePendingUploadInput struct {
	ID                 uuid.UUID
	FeedbackID         string
	UserID             string
	ObjectKey          string
	OriginalFileName   string
	ContentType        string
	RequestedSizeBytes int64
}

type MarkUploadedInput struct {
	ObjectKey   string
	FeedbackID  string
	UserID      string
	SizeBytes   int64
	ContentType string
}

type FindUploadedByObjectKeyInput struct {
	ObjectKey  string
	FeedbackID string
	UserID     string
}

type ListUploadedByFeedbackIDInput struct {
	FeedbackID string
	UserID     string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreatePendingUpload(
	ctx context.Context,
	input CreatePendingUploadInput,
) (*Attachment, error) {
	const query = `
		INSERT INTO feedback_attachments (
			id,
			feedback_id,
			user_id,
			object_key,
			original_file_name,
			content_type,
			requested_size_bytes,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			id,
			feedback_id,
			user_id,
			object_key,
			original_file_name,
			content_type,
			requested_size_bytes,
			size_bytes,
			status,
			created_at,
			uploaded_at,
			deleted_at,
			updated_at
	`

	var attachment Attachment

	err := r.db.QueryRow(
		ctx,
		query,
		input.ID,
		input.FeedbackID,
		input.UserID,
		input.ObjectKey,
		input.OriginalFileName,
		input.ContentType,
		input.RequestedSizeBytes,
		AttachmentStatusPendingUpload,
	).Scan(
		&attachment.ID,
		&attachment.FeedbackID,
		&attachment.UserID,
		&attachment.ObjectKey,
		&attachment.OriginalFileName,
		&attachment.ContentType,
		&attachment.RequestedSizeBytes,
		&attachment.SizeBytes,
		&attachment.Status,
		&attachment.CreatedAt,
		&attachment.UploadedAt,
		&attachment.DeletedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create pending upload attachment: %w", err)
	}

	return &attachment, nil
}

func (r *Repository) MarkUploaded(
	ctx context.Context,
	input MarkUploadedInput,
) (*Attachment, error) {
	const query = `
		UPDATE feedback_attachments
		SET
			status = $1,
			size_bytes = $2,
			content_type = $3,
			uploaded_at = NOW(),
			updated_at = NOW()
		WHERE
			object_key = $4
			AND feedback_id = $5
			AND user_id = $6
			AND status = $7
		RETURNING
			id,
			feedback_id,
			user_id,
			object_key,
			original_file_name,
			content_type,
			requested_size_bytes,
			size_bytes,
			status,
			created_at,
			uploaded_at,
			deleted_at,
			updated_at
	`

	var attachment Attachment

	err := r.db.QueryRow(
		ctx,
		query,
		AttachmentStatusUploaded,
		input.SizeBytes,
		input.ContentType,
		input.ObjectKey,
		input.FeedbackID,
		input.UserID,
		AttachmentStatusPendingUpload,
	).Scan(
		&attachment.ID,
		&attachment.FeedbackID,
		&attachment.UserID,
		&attachment.ObjectKey,
		&attachment.OriginalFileName,
		&attachment.ContentType,
		&attachment.RequestedSizeBytes,
		&attachment.SizeBytes,
		&attachment.Status,
		&attachment.CreatedAt,
		&attachment.UploadedAt,
		&attachment.DeletedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttachmentNotPendingUpload
		}

		return nil, fmt.Errorf("mark attachment uploaded: %w", err)
	}

	return &attachment, nil
}

func (r *Repository) FindUploadedByObjectKey(
	ctx context.Context,
	input FindUploadedByObjectKeyInput,
) (*Attachment, error) {
	const query = `
		SELECT
			id,
			feedback_id,
			user_id,
			object_key,
			original_file_name,
			content_type,
			requested_size_bytes,
			size_bytes,
			status,
			created_at,
			uploaded_at,
			deleted_at,
			updated_at
		FROM feedback_attachments
		WHERE
			object_key = $1
			AND feedback_id = $2
			AND user_id = $3
			AND status = $4
	`

	var attachment Attachment

	err := r.db.QueryRow(
		ctx,
		query,
		input.ObjectKey,
		input.FeedbackID,
		input.UserID,
		AttachmentStatusUploaded,
	).Scan(
		&attachment.ID,
		&attachment.FeedbackID,
		&attachment.UserID,
		&attachment.ObjectKey,
		&attachment.OriginalFileName,
		&attachment.ContentType,
		&attachment.RequestedSizeBytes,
		&attachment.SizeBytes,
		&attachment.Status,
		&attachment.CreatedAt,
		&attachment.UploadedAt,
		&attachment.DeletedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttachmentNotUploaded
		}

		return nil, fmt.Errorf("find uploaded attachment by object key: %w", err)
	}

	return &attachment, nil
}

func (r *Repository) ListUploadedByFeedbackID(
	ctx context.Context,
	input ListUploadedByFeedbackIDInput,
) ([]Attachment, error) {
	const query = `
		SELECT
			id,
			feedback_id,
			user_id,
			object_key,
			original_file_name,
			content_type,
			requested_size_bytes,
			size_bytes,
			status,
			created_at,
			uploaded_at,
			deleted_at,
			updated_at
		FROM feedback_attachments
		WHERE
			feedback_id = $1
			AND user_id = $2
			AND status = $3
		ORDER BY uploaded_at DESC, created_at DESC
	`

	rows, err := r.db.Query(
		ctx,
		query,
		input.FeedbackID,
		input.UserID,
		AttachmentStatusUploaded,
	)
	if err != nil {
		return nil, fmt.Errorf("list uploaded attachments by feedback id: %w", err)
	}
	defer rows.Close()

	attachments := make([]Attachment, 0)

	for rows.Next() {
		var attachment Attachment

		err := rows.Scan(
			&attachment.ID,
			&attachment.FeedbackID,
			&attachment.UserID,
			&attachment.ObjectKey,
			&attachment.OriginalFileName,
			&attachment.ContentType,
			&attachment.RequestedSizeBytes,
			&attachment.SizeBytes,
			&attachment.Status,
			&attachment.CreatedAt,
			&attachment.UploadedAt,
			&attachment.DeletedAt,
			&attachment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan uploaded attachment: %w", err)
		}

		attachments = append(attachments, attachment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate uploaded attachments: %w", err)
	}

	return attachments, nil
}
