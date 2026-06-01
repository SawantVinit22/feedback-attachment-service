output "policy_arn" {
  description = "ARN of the backend S3 access IAM policy"
  value       = aws_iam_policy.backend_s3_access.arn
}

output "policy_name" {
  description = "Name of the backend S3 access IAM policy"
  value       = aws_iam_policy.backend_s3_access.name
}
