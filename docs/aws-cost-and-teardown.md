# AWS Cost Control and Teardown Runbook

## Purpose

This project uses AWS resources for learning and development. The goal is to build production-style architecture while keeping cloud cost controlled.

All billable resources must be created through Terraform wherever possible so they can be destroyed cleanly when not needed.

## Core Rule

```text
If Terraform creates it, Terraform must be able to destroy it.
```

Avoid creating long-running AWS resources manually from the console unless they are temporary and documented.

## Current AWS Resources

The project currently uses:

```text
S3 bucket for attachment storage
IAM policy for backend S3 access
IAM user for local backend testing
Terraform-managed infrastructure
```

## Resources Safe to Keep During Learning

These resources are generally low-cost for small-scale development:

```text
IAM users, policies, and roles
Small S3 bucket with minimal test files
Terraform code
Local Docker image
Local AWS CLI profiles
```

Even though S3 is usually low-cost at small scale, test files should still be cleaned up regularly.

## Resources That Should Not Run Continuously

Avoid keeping these resources running when not actively testing:

```text
RDS PostgreSQL
NAT Gateway
Application Load Balancer
ECS services
EKS clusters
EC2 instances
Aurora databases
```

These can create continuous charges.

## Recommended AWS Direction

For this project, the preferred AWS-only architecture is:

```text
Amazon S3              -> actual file storage
Amazon RDS PostgreSQL  -> attachment and feedback metadata
Amazon ECR             -> Docker image registry
Amazon ECS             -> first deployment target
Terraform              -> infrastructure lifecycle
```

EKS should be avoided in the early learning phase because it adds cost and operational complexity.

## Database Direction

For attachment metadata, prefer Amazon RDS PostgreSQL over DynamoDB because the attachment module will eventually merge with a larger feedback service.

The future feedback system may contain relational entities such as:

```text
users
feedback
attachments
comments
projects or organizations
audit logs
```

A relational database is better for these relationships, joins, indexing, constraints, and future reporting queries.

## Future Attachment Metadata Table

Expected table:

```text
feedback_attachments
```

Expected columns:

```text
id
feedback_id
user_id
object_key
original_file_name
content_type
size_bytes
status
created_at
uploaded_at
deleted_at
```

Expected statuses:

```text
PENDING_UPLOAD
UPLOADED
FAILED
DELETED
```

## Development RDS Rules

For dev-only RDS PostgreSQL, use teardown-friendly settings:

```hcl
deletion_protection = false
skip_final_snapshot = true
```

Reason:

```text
deletion_protection = false  -> Terraform can destroy the DB
skip_final_snapshot = true   -> avoids leftover snapshot cost in dev
```

For production, these values should be different:

```hcl
deletion_protection = true
skip_final_snapshot = false
```

## RDS Usage Rule

For learning:

```text
Short break       -> stop RDS if needed
Done testing      -> terraform destroy
Long idle period  -> destroy RDS completely
```

Stopping RDS can reduce compute cost, but storage and snapshots may still cost money. For learning environments, destruction is safer when the DB is not needed.

## S3 Cleanup Rule

Before destroying an S3 bucket, the bucket must be empty.

If Terraform destroy fails because the bucket is not empty, delete test objects first.

Example cleanup command:

```bash
aws s3 rm s3://your-s3-bucket-name --recursive --profile your-aws-profile
```

Then run Terraform destroy again.

## Terraform Destroy

From the dev environment folder:

```bash
cd infra/environments/dev
terraform plan -destroy
terraform destroy
```

Always inspect the destroy plan before confirming.

## Terraform Safety Checklist

Before running destroy:

```text
Confirm you are using the dev AWS account/profile
Confirm you are in infra/environments/dev
Confirm no production resources are referenced
Confirm S3 test files can be deleted
Confirm RDS final snapshot settings are dev-safe
```

## AWS Budget Recommendation

Before adding RDS, ECS, or any always-running service, create an AWS monthly budget alert.

Recommended learning budget:

```text
Monthly budget: 5 to 10 USD
Alert thresholds: 50%, 80%, 100%
Notification: email
```

## Resources to Avoid for Now

Do not add these yet:

```text
EKS cluster
NAT Gateway
Aurora
Multi-AZ RDS
Large EC2 instances
Always-running ECS service
Production Load Balancer
```

These should be introduced only when there is a clear learning or deployment goal.

## Local Development Preference

Prefer local Docker for application runtime testing.

Prefer AWS only when testing actual AWS integration such as:

```text
S3 presigned URLs
IAM permissions
RDS connectivity
ECR image push
ECS deployment
```

## Teardown Mindset

Every new AWS milestone should answer:

```text
What resources will this create?
Will this cost money while idle?
How do we destroy it?
Will anything remain after destroy?
Does the README/runbook explain cleanup?
```

## Next Planned AWS Milestones

Recommended sequence:

```text
1. Add S3 lifecycle cleanup for dev uploads
2. Add RDS PostgreSQL metadata design
3. Add RDS Terraform module
4. Add DB migration files
5. Add Go repository layer
6. Update APIs to persist attachment metadata
7. Add ECR only after Docker image flow is ready
8. Add ECS deployment only after app and DB are stable
```
