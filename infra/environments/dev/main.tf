terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

module "storage" {
  source = "../../modules/storage"

  project_name    = var.project_name
  environment     = var.environment
  bucket_name     = var.bucket_name
  allowed_origins = var.allowed_origins
}

module "backend_iam" {
  source = "../../modules/backend_iam"

  project_name = var.project_name
  environment  = var.environment
  bucket_arn   = module.storage.bucket_arn
}

module "local_backend_user" {
  source = "../../modules/local_backend_user"

  project_name          = var.project_name
  environment           = var.environment
  backend_s3_policy_arn = module.backend_iam.policy_arn
}