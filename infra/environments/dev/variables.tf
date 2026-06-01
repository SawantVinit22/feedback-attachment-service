variable "aws_region" {
  type        = string
  description = "AWS region where resources will be created"
}

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "environment" {
  type        = string
  description = "Environment name"
}

variable "bucket_name" {
  type        = string
  description = "Globally unique S3 bucket name"
}

variable "allowed_origins" {
  type        = list(string)
  description = "Allowed frontend origins"
}
