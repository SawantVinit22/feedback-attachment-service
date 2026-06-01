variable "project_name" {
  type        = string
  description = "Project name used for naming and tagging IAM resources"
}

variable "environment" {
  type        = string
  description = "Environment name like dev, test, prod"
}

variable "bucket_arn" {
  type        = string
  description = "ARN of the S3 bucket used for attachment storage"
}
