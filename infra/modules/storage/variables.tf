variable "project_name" {
  type        = string
  description = "Project name used for naming and tagging resources"
}

variable "environment" {
  type        = string
  description = "Environment name like dev, test, prod"
}

variable "bucket_name" {
  type        = string
  description = "Globally unique S3 bucket name"
}

variable "allowed_origins" {
  type        = list(string)
  description = "Allowed frontend origins for browser uploads"
  default     = []
}
