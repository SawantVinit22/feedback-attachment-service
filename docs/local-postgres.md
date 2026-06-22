# Local PostgreSQL Runbook

## Purpose

This project uses PostgreSQL for future attachment metadata persistence.

For local development, PostgreSQL runs through Docker Compose so the application can be tested without creating Amazon RDS resources.

## Start PostgreSQL

From the repository root:

```bash
docker compose up -d postgres
```

## Check Container Status

```bash
docker compose ps
```

Expected container:

```text
feedback-attachment-postgres
```

## View Logs

```bash
docker compose logs -f postgres
```

## Connect Using psql Inside Container

```bash
docker exec -it feedback-attachment-postgres psql -U feedback_app -d feedback_attachments
```

Inside psql, test:

```sql
SELECT current_database();
SELECT current_user;
```

Exit psql:

```text
\q
```

## Database URL

Local development connection string:

```text
postgres://feedback_app:feedback_app_password@localhost:5432/feedback_attachments?sslmode=disable
```

Use this when the Go application later reads `DATABASE_URL`.

## Stop PostgreSQL

```bash
docker compose stop postgres
```

## Start Again

```bash
docker compose start postgres
```

## Destroy Local PostgreSQL Data

This removes the container and the local database volume.

Use only when you are okay losing local data.

```bash
docker compose down -v
```

## Cost Note

This local PostgreSQL setup does not create AWS resources.

Use this before introducing Amazon RDS PostgreSQL.
