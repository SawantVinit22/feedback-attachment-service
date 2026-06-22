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
# Feedback Attachment Service

A Go backend service for securely managing feedback attachments using AWS S3 presigned URLs and PostgreSQL-backed attachment metadata.

The service supports direct client uploads to S3 while maintaining a backend-controlled attachment lifecycle in PostgreSQL.

## Current Features

* Private S3 bucket managed using Terraform
* Backend IAM user with limited S3 permissions
* S3 presigned upload URL generation
* Direct client upload to S3
* PostgreSQL metadata persistence
* Attachment lifecycle tracking
* Upload completion verification using S3 `HeadObject`
* Presigned download URL generation
* Download validation using uploaded metadata
* List uploaded attachments by `feedback_id`
* Clean lifecycle-safe API errors
* Dockerfile for containerized service build
* Docker Compose setup for local PostgreSQL
* Makefile for local development commands
* Refactored Go project structure

## Attachment Lifecycle

```text
PENDING_UPLOAD -> UPLOADED -> DOWNLOAD_ALLOWED / LIST_VISIBLE
```

### Lifecycle Flow

```text
Client
  -> POST /attachments/presign-upload
  -> receives attachment_id + upload_url + object_key + required_headers
  -> uploads file directly to S3 using PUT
  -> POST /attachments/complete-upload
  -> backend verifies S3 object using HeadObject
  -> backend updates PostgreSQL row from PENDING_UPLOAD to UPLOADED
  -> POST /attachments/presign-download
  -> backend validates uploaded metadata
  -> receives download_url
  -> downloads file directly from S3
  -> GET /attachments?feedback_id=<feedback_id>
  -> receives uploaded attachment metadata
```

## Project Structure

```text
.
├── app/
│   ├── cmd/server/main.go
│   ├── internal/config/config.go
│   ├── internal/database/database.go
│   ├── internal/attachments/
│   │   ├── models.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── validation.go
│   └── internal/httpapi/
│       ├── handlers.go
│       └── response.go
├── infra/
│   ├── environments/dev/
│   └── modules/
├── migrations/
│   ├── 001_create_feedback_attachments.up.sql
│   └── 001_create_feedback_attachments.down.sql
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── .env.example
└── README.md
```

## Local Development

### 1. Create Local Environment File

Create a local `.env` file from `.env.example`.

```bash
cp .env.example .env
```

Example values:

```env
SERVER_ADDR=:8080
AWS_PROFILE=your-aws-profile
AWS_REGION=us-west-2
S3_BUCKET_NAME=your-s3-bucket-name
DATABASE_URL=postgres://feedback_app:feedback_app_password@localhost:5432/feedback_attachments?sslmode=disable
```

Do not commit `.env`.

### 2. Start PostgreSQL

```bash
make postgres-up
```

### 3. Apply Database Migration

```bash
docker exec -i feedback-attachment-postgres \
  psql -U feedback_app -d feedback_attachments \
  < migrations/001_create_feedback_attachments.up.sql
```

### 4. Run the Service

```bash
make run
```

Expected logs:

```text
connected to postgres
server started on :8080
```

### 5. Health Check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "ok"
}
```

## API Endpoints

### 1. Presign Upload

```http
POST /attachments/presign-upload
```

Creates a `PENDING_UPLOAD` row in PostgreSQL and returns a presigned S3 upload URL.

Request:

```bash
curl -s -X POST http://localhost:8080/attachments/presign-upload \
  -H "Content-Type: application/json" \
  -d '{
    "feedback_id": "fb_001",
    "file_name": "sample.txt",
    "content_type": "text/plain",
    "size_bytes": 100
  }' | jq
```

Example response:

```json
{
  "attachment_id": "00000000-0000-0000-0000-000000000000",
  "upload_url": "https://...",
  "object_key": "feedback/fb_001/attachments/00000000-0000-0000-0000-000000000000/sample.txt",
  "expires_in_seconds": 600,
  "required_headers": {
    "Content-Type": "text/plain",
    "x-amz-meta-feedback-id": "fb_001",
    "x-amz-meta-original-file-name": "sample.txt",
    "x-amz-meta-uploaded-by": "user_123"
  }
}
```

### 2. Upload File to S3

The client uploads the file directly to S3 using the returned `upload_url`.

All required signed headers must be included.

```bash
echo "sample attachment content" > sample.txt

curl -X PUT \
  -H "Content-Type: text/plain" \
  -H "x-amz-meta-feedback-id: fb_001" \
  -H "x-amz-meta-original-file-name: sample.txt" \
  -H "x-amz-meta-uploaded-by: user_123" \
  --upload-file sample.txt \
  "$UPLOAD_URL"
```

### 3. Complete Upload

```http
POST /attachments/complete-upload
```

Verifies the uploaded object in S3 and updates the PostgreSQL row from `PENDING_UPLOAD` to `UPLOADED`.

Request:

```bash
curl -s -X POST http://localhost:8080/attachments/complete-upload \
  -H "Content-Type: application/json" \
  -d "{
    \"feedback_id\": \"fb_001\",
    \"object_key\": \"$OBJECT_KEY\"
  }" | jq
```

Example response:

```json
{
  "attachment_id": "00000000-0000-0000-0000-000000000000",
  "object_key": "feedback/fb_001/attachments/00000000-0000-0000-0000-000000000000/sample.txt",
  "status": "UPLOADED",
  "size_bytes": 25,
  "content_type": "text/plain",
  "metadata": {
    "feedback-id": "fb_001",
    "original-file-name": "sample.txt",
    "uploaded-by": "user_123"
  }
}
```

### 4. Presign Download

```http
POST /attachments/presign-download
```

Generates a download URL only if the attachment exists in PostgreSQL with status `UPLOADED`.

Request:

```bash
curl -s -X POST http://localhost:8080/attachments/presign-download \
  -H "Content-Type: application/json" \
  -d "{
    \"feedback_id\": \"fb_001\",
    \"object_key\": \"$OBJECT_KEY\"
  }" | jq
```

Example response:

```json
{
  "attachment_id": "00000000-0000-0000-0000-000000000000",
  "download_url": "https://...",
  "object_key": "feedback/fb_001/attachments/00000000-0000-0000-0000-000000000000/sample.txt",
  "expires_in_seconds": 600,
  "size_bytes": 25,
  "content_type": "text/plain"
}
```

### 5. Download File

```bash
curl -L "$DOWNLOAD_URL" -o downloaded-sample.txt
```

### 6. List Uploaded Attachments

```http
GET /attachments?feedback_id=<feedback_id>
```

Returns uploaded attachments for the current user and feedback item.

Request:

```bash
curl -s "http://localhost:8080/attachments?feedback_id=fb_001" | jq
```

Example response:

```json
[
  {
    "attachment_id": "00000000-0000-0000-0000-000000000000",
    "feedback_id": "fb_001",
    "object_key": "feedback/fb_001/attachments/00000000-0000-0000-0000-000000000000/sample.txt",
    "original_file_name": "sample.txt",
    "content_type": "text/plain",
    "size_bytes": 25,
    "status": "UPLOADED",
    "uploaded_at": "2026-06-22T18:24:43.556468+05:30"
  }
]
```

## Clean Error Responses

If a download is requested for an attachment that is not uploaded, the API returns:

```json
{
  "error": "attachment not found or not uploaded"
}
```

If upload completion is attempted for an attachment that is not pending upload, the API returns:

```json
{
  "error": "attachment not found or not pending upload"
}
```

## Makefile Commands

Start PostgreSQL:

```bash
make postgres-up
```

Stop PostgreSQL:

```bash
make postgres-down
```

View PostgreSQL logs:

```bash
make postgres-logs
```

Run the service locally:

```bash
make run
```

Build the Go app:

```bash
make build
```

Run tests:

```bash
make test
```

## Docker

Build the image:

```bash
docker build -t feedback-attachment-service:local .
```

Run the image locally:

```bash
docker run --rm \
  --name feedback-attachment-service \
  -p 8080:8080 \
  -e AWS_PROFILE=your-aws-profile \
  -e AWS_REGION=us-west-2 \
  -e S3_BUCKET_NAME=your-s3-bucket-name \
  -e DATABASE_URL=postgres://feedback_app:feedback_app_password@host.docker.internal:5432/feedback_attachments?sslmode=disable \
  -e HOME=/home/app \
  -v "$HOME/.aws:/home/app/.aws:ro" \
  feedback-attachment-service:local
```

## Database Table

The service stores attachment metadata in PostgreSQL table:

```text
feedback_attachments
```

Important fields:

```text
id
feedback_id
user_id
object_key
original_file_name
content_type
requested_size_bytes
size_bytes
status
created_at
uploaded_at
deleted_at
updated_at
```

Supported statuses:

```text
PENDING_UPLOAD
UPLOADED
FAILED
DELETED
```

## Current Limitations

* User ID is currently hardcoded as `user_123`
* No real authentication yet
* No soft delete API yet
* No background cleanup job for stale `PENDING_UPLOAD` records yet
* No CI/CD pipeline yet
* No production deployment setup yet

## Roadmap

1. Add authentication and authorization
2. Add soft delete attachment API
3. Add cleanup job for stale pending uploads
4. Add automated tests
5. Add CI/CD pipeline
6. Add production deployment infrastructure
7. Move from local PostgreSQL to managed PostgreSQL for deployed environments
