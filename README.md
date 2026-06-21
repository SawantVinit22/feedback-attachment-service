# Feedback Attachment Service

A Go backend service for generating secure AWS S3 presigned URLs for uploading and downloading feedback attachments.

## Current Features

- Private S3 bucket managed by Terraform
- Backend IAM user with limited S3 permissions
- Presigned upload URL generation
- Direct client upload to S3
- Upload completion verification using S3 HeadObject
- Presigned download URL generation
- Basic feedback/user ownership validation
- Refactored Go project structure

## API Flow

```text
Client
  → POST /attachments/presign-upload
  → receives upload_url + object_key + required_headers
  → uploads file directly to S3 using PUT
  → POST /attachments/complete-upload
  → backend verifies object using HeadObject
  → POST /attachments/presign-download
  → receives download_url
  → downloads file directly from S3
```

## Project Structure

```text
app/
├── cmd/server/main.go
├── internal/config/config.go
├── internal/attachments/
│   ├── models.go
│   ├── service.go
│   └── validation.go
└── internal/httpapi/
    ├── handlers.go
    └── response.go
```

## Required Environment Variables

```bash
AWS_PROFILE=your-aws-profile
AWS_REGION=us-west-2
S3_BUCKET_NAME=your-s3-bucket-name
```

Optional:

```bash
SERVER_ADDR=:8080
```

## Run Locally

```bash
cd app

AWS_PROFILE=your-aws-profile \
AWS_REGION=us-west-2 \
S3_BUCKET_NAME=your-s3-bucket-name \
go run ./cmd/server
```

## Health Check

```bash
curl http://localhost:8080/health
```

Expected:

```json
{"status":"ok"}
```

## Presign Upload

```bash
curl -s -X POST http://localhost:8080/attachments/presign-upload \
  -H "Content-Type: application/json" \
  -d '{
    "feedback_id": "fb_456",
    "file_name": "sample.txt",
    "content_type": "text/plain",
    "size_bytes": 100
  }' | jq
```

## Upload File to S3

Use the returned `upload_url` and `required_headers`.

```bash
curl -X PUT \
  -H "Content-Type: text/plain" \
  -H "x-amz-meta-feedback-id: fb_456" \
  -H "x-amz-meta-original-file-name: sample.txt" \
  -H "x-amz-meta-uploaded-by: user_123" \
  --upload-file sample.txt \
  "$UPLOAD_URL"
```

## Complete Upload

```bash
curl -s -X POST http://localhost:8080/attachments/complete-upload \
  -H "Content-Type: application/json" \
  -d "{
    \"feedback_id\": \"fb_456\",
    \"object_key\": \"$OBJECT_KEY\"
  }" | jq
```

## Presign Download

```bash
curl -s -X POST http://localhost:8080/attachments/presign-download \
  -H "Content-Type: application/json" \
  -d "{
    \"feedback_id\": \"fb_456\",
    \"object_key\": \"$OBJECT_KEY\"
  }" | jq
```

## Download File

```bash
curl -L "$DOWNLOAD_URL" -o downloaded-sample.txt
```

## Current Limitations

- User ID is currently hardcoded as `user_123`
- No real authentication yet
- No metadata database yet
- No attachment status persistence yet
- No Dockerfile yet
- No CI/CD pipeline yet
- No production deployment setup yet

## Next Roadmap

1. Add Dockerfile
2. Add metadata database
3. Add authentication and authorization
4. Add CI/CD pipeline
5. Add production deployment infrastructure
