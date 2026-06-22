ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: run build test postgres-up postgres-down postgres-logs

run:
	cd app && go run ./cmd/server

build:
	cd app && go build ./...

test:
	cd app && go test ./...

postgres-up:
	docker compose up -d postgres

postgres-down:
	docker compose down

postgres-logs:
	docker compose logs -f postgres