output "user_name" {
  description = "Name of the local backend IAM user"
  value       = aws_iam_user.local_backend.name
}

output "user_arn" {
  description = "ARN of the local backend IAM user"
  value       = aws_iam_user.local_backend.arn
}
