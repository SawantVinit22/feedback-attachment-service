CREATE TABLE IF NOT EXISTS feedback_attachments (
    id UUID PRIMARY KEY,
    feedback_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    original_file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,

    requested_size_bytes BIGINT NOT NULL CHECK (requested_size_bytes > 0),
    size_bytes BIGINT CHECK (size_bytes IS NULL OR size_bytes >= 0),

    status TEXT NOT NULL CHECK (
        status IN (
            'PENDING_UPLOAD',
            'UPLOADED',
            'FAILED',
            'DELETED'
        )
    ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uploaded_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedback_attachments_feedback_id
    ON feedback_attachments (feedback_id);

CREATE INDEX IF NOT EXISTS idx_feedback_attachments_user_id
    ON feedback_attachments (user_id);

CREATE INDEX IF NOT EXISTS idx_feedback_attachments_status
    ON feedback_attachments (status);

CREATE INDEX IF NOT EXISTS idx_feedback_attachments_feedback_status
    ON feedback_attachments (feedback_id, status);

CREATE INDEX IF NOT EXISTS idx_feedback_attachments_created_at
    ON feedback_attachments (created_at);