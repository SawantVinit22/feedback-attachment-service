# Attachment Metadata and Lifecycle Design

## Purpose

This document defines how feedback attachments should be tracked, verified, downloaded, deleted, and cleaned up safely.

The service stores file bytes in Amazon S3, but attachment ownership and lifecycle state should be stored in a relational database.

## Core Principle

S3 stores file bytes. PostgreSQL stores attachment metadata and lifecycle state. The application controls transitions between states.

S3 lifecycle rules should not blindly delete real feedback attachments because S3 is not aware of database state.

## Recommended Database

Use Amazon RDS PostgreSQL.

Reason: this attachment module will later merge with a larger feedback service where feedback, users, comments, attachments, and audit logs are relational entities.

PostgreSQL is better for joins, indexes, constraints, reporting queries, and future relational growth.

## Table: feedback_attachments

Expected columns:

- id
- feedback_id
- user_id
- object_key
- original_file_name
- content_type
- size_bytes
- status
- created_at
- uploaded_at
- deleted_at

## Status Values

- PENDING_UPLOAD
- UPLOADED
- FAILED
- DELETED

## Status Meaning

### PENDING_UPLOAD

The backend generated a presigned upload URL, but the file has not yet been verified in S3.

Possible cases:

- Client has not uploaded yet
- Client upload is still in progress
- Client abandoned the upload
- Upload failed before completion

### UPLOADED

The backend verified the object exists in S3 using HeadObject and confirmed ownership metadata.

This is a real attachment linked to feedback.

### FAILED

The upload was not completed or verification failed.

This state can be used for expired or abandoned upload attempts.

### DELETED

The attachment was deleted by user action, feedback deletion, retention policy, or admin/system cleanup.

## Object Key Strategy

Current object key pattern:

feedback/{feedback_id}/users/{user_id}/{uuid}-{file_name}

This is acceptable for the current learning phase.

Later, when database metadata is added, the preferred object key should include attachment ID:

feedback/{feedback_id}/attachments/{attachment_id}/{file_name}

Reason:

- attachment_id becomes the stable unique identity
- object key is easier to trace
- user_id can change or become less important in the storage path
- feedback ownership remains clear

## Upload Flow

POST /attachments/presign-upload

- validate request
- create attachment_id
- build object_key
- insert DB record with PENDING_UPLOAD
- generate S3 presigned PUT URL
- return attachment_id, object_key, upload_url, required_headers

## Complete Upload Flow

POST /attachments/complete-upload

- receive attachment_id and object_key
- load DB record
- verify user/feedback ownership
- call S3 HeadObject
- verify S3 metadata
- update DB record to UPLOADED
- store actual size, content type, and uploaded_at

## Download Flow

POST /attachments/presign-download

- receive attachment_id or object_key
- load DB record
- verify status is UPLOADED
- verify user/feedback ownership
- generate S3 presigned GET URL

## Deletion Flow

DELETE /attachments/{attachment_id}

- load DB record
- verify user/feedback ownership
- delete S3 object or mark for async deletion
- update DB status to DELETED
- set deleted_at

## Cleanup Rules

Do not blindly delete objects under feedback/ because they may be valid feedback attachments.

Safe cleanup should be DB-aware.

Allowed cleanup:

- PENDING_UPLOAD records older than defined expiry window
- FAILED records after retention period
- DELETED records after deletion is confirmed
- temporary upload prefixes if introduced later

Not allowed cleanup:

- UPLOADED attachments still linked to active feedback
- objects under feedback/ without checking DB state
- objects only based on age without checking lifecycle status

## Stale Pending Upload Cleanup

Future cleanup job:

- find PENDING_UPLOAD records older than 24 hours
- check whether S3 object exists
- delete S3 object if present
- mark DB record as FAILED or EXPIRED

This avoids orphaned files and broken database references.

## S3 Lifecycle Rules

S3 lifecycle rules may be used only for clearly disposable prefixes, such as:

- tmp/
- pending/
- multipart/

S3 lifecycle should not be applied to permanent feedback attachments unless a business retention rule is defined.

## Future APIs

Potential future APIs:

- GET /feedback/{feedback_id}/attachments
- GET /attachments/{attachment_id}
- DELETE /attachments/{attachment_id}
- POST /attachments/{attachment_id}/presign-download

## Future Indexes

Expected PostgreSQL indexes:

- primary key on id
- unique index on object_key
- index on feedback_id
- index on user_id
- index on status
- index on feedback_id + status
- index on created_at

## Cost and Teardown Consideration

RDS PostgreSQL should be introduced only when needed and managed through Terraform.

For dev environments:

deletion_protection = false
skip_final_snapshot = true

This keeps teardown simple and avoids unnecessary learning-environment cost.