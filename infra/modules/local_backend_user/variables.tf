variable "project_name" {
  type        = string
  description = "Project name used for naming and tagging IAM resources"
}

variable "environment" {
  type        = string
  description = "Environment name like dev, test, prod"
}

variable "backend_s3_policy_arn" {
  type        = string
  description = "ARN of the backend S3 access policy to attach to the local backend user"
}
