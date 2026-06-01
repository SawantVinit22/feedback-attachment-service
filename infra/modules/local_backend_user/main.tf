resource "aws_iam_user" "local_backend" {
  name = "${var.project_name}-${var.environment}-local-backend-user"

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Purpose     = "local-backend-testing"
  }
}

resource "aws_iam_user_policy_attachment" "backend_s3_access" {
  user       = aws_iam_user.local_backend.name
  policy_arn = var.backend_s3_policy_arn
}
