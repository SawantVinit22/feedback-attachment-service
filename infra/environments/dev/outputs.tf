output "bucket_name" {
  description = "Created S3 bucket name"
  value       = module.storage.bucket_name
}

output "bucket_arn" {
  description = "Created S3 bucket ARN"
  value       = module.storage.bucket_arn
}

output "backend_s3_policy_arn" {
  description = "ARN of the backend S3 access IAM policy"
  value       = module.backend_iam.policy_arn
}

output "backend_s3_policy_name" {
  description = "Name of the backend S3 access IAM policy"
  value       = module.backend_iam.policy_name
}
