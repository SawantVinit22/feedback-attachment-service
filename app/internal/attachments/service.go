package attachments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const presignExpiry = 10 * time.Minute

type S3PresignService struct {
	bucket     string
	client     *s3.Client
	presigner  *s3.PresignClient
	repository *Repository
}

func NewS3PresignService(
	ctx context.Context,
	bucket string,
	repository *Repository,
) (*S3PresignService, error) {
	if strings.TrimSpace(bucket) == "" {
		return nil, errors.New("S3_BUCKET_NAME is required")
	}

	if repository == nil {
		return nil, errors.New("attachment repository is required")
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &S3PresignService{
		bucket:     bucket,
		client:     s3Client,
		presigner:  s3.NewPresignClient(s3Client),
		repository: repository,
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

	attachmentID := uuid.New()
	objectKey := buildObjectKey(req.FeedbackID, attachmentID.String(), req.FileName)

	_, err := s.repository.CreatePendingUpload(ctx, CreatePendingUploadInput{
		ID:                 attachmentID,
		FeedbackID:         req.FeedbackID,
		UserID:             userID,
		ObjectKey:          objectKey,
		OriginalFileName:   req.FileName,
		ContentType:        req.ContentType,
		RequestedSizeBytes: req.SizeBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("create pending upload record: %w", err)
	}

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
		AttachmentID:     attachmentID.String(),
		UploadURL:        presignedReq.URL,
		ObjectKey:        objectKey,
		ExpiresInSeconds: int64(presignExpiry.Seconds()),
		RequiredHeaders: map[string]string{
			"Content-Type":                  req.ContentType,
			"x-amz-meta-feedback-id":        req.FeedbackID,
			"x-amz-meta-original-file-name": req.FileName,
			"x-amz-meta-uploaded-by":        userID,
		},
	}, nil
}

func (s *S3PresignService) CompleteUpload(
	ctx context.Context,
	req CompleteUploadRequest,
	userID string,
) (*CompleteUploadResponse, error) {
	objectKey := strings.TrimSpace(req.ObjectKey)

	headOutput, err := s.verifyObjectOwnership(ctx, req.FeedbackID, objectKey, userID)
	if err != nil {
		return nil, err
	}

	actualSizeBytes := aws.ToInt64(headOutput.ContentLength)
	actualContentType := aws.ToString(headOutput.ContentType)

	attachment, err := s.repository.MarkUploaded(ctx, MarkUploadedInput{
		ObjectKey:   objectKey,
		FeedbackID:  req.FeedbackID,
		UserID:      userID,
		SizeBytes:   actualSizeBytes,
		ContentType: actualContentType,
	})
	if err != nil {
		if errors.Is(err, ErrAttachmentNotPendingUpload) {
			return nil, err
		}

		return nil, fmt.Errorf("update upload metadata: %w", err)
	}

	return &CompleteUploadResponse{
		AttachmentID: attachment.ID.String(),
		ObjectKey:    objectKey,
		Status:       string(attachment.Status),
		SizeBytes:    actualSizeBytes,
		ContentType:  actualContentType,
		Metadata:     headOutput.Metadata,
	}, nil
}

func (s *S3PresignService) GenerateDownloadURL(
	ctx context.Context,
	req PresignDownloadRequest,
	userID string,
) (*PresignDownloadResponse, error) {
	objectKey := strings.TrimSpace(req.ObjectKey)

	attachment, err := s.repository.FindUploadedByObjectKey(ctx, FindUploadedByObjectKeyInput{
		ObjectKey:  objectKey,
		FeedbackID: req.FeedbackID,
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, ErrAttachmentNotUploaded) {
			return nil, err
		}

		return nil, fmt.Errorf("validate uploaded attachment metadata: %w", err)
	}

	headOutput, err := s.verifyObjectOwnership(ctx, req.FeedbackID, objectKey, userID)
	if err != nil {
		return nil, err
	}

	presignedReq, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = presignExpiry
	})
	if err != nil {
		return nil, fmt.Errorf("generate download presigned url: %w", err)
	}

	return &PresignDownloadResponse{
		AttachmentID:     attachment.ID.String(),
		DownloadURL:      presignedReq.URL,
		ObjectKey:        objectKey,
		ExpiresInSeconds: int64(presignExpiry.Seconds()),
		SizeBytes:        aws.ToInt64(headOutput.ContentLength),
		ContentType:      aws.ToString(headOutput.ContentType),
	}, nil
}

func (s *S3PresignService) verifyObjectOwnership(
	ctx context.Context,
	feedbackID string,
	objectKey string,
	userID string,
) (*s3.HeadObjectOutput, error) {
	feedbackID = strings.TrimSpace(feedbackID)
	if feedbackID == "" {
		return nil, errors.New("feedback_id is required")
	}

	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return nil, errors.New("object_key is required")
	}

	legacyUserPrefix := fmt.Sprintf("feedback/%s/users/%s/", feedbackID, userID)
	attachmentPrefix := fmt.Sprintf("feedback/%s/attachments/", feedbackID)

	if !strings.HasPrefix(objectKey, legacyUserPrefix) &&
		!strings.HasPrefix(objectKey, attachmentPrefix) {
		return nil, errors.New("object_key does not belong to this feedback/user")
	}

	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("verify object: %w", err)
	}

	metadata := headOutput.Metadata

	if metadata["uploaded-by"] != userID {
		return nil, errors.New("uploaded object does not belong to current user")
	}

	if metadata["feedback-id"] != feedbackID {
		return nil, errors.New("uploaded object does not belong to this feedback")
	}

	return headOutput, nil
}
