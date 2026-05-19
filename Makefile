DATABASE_URL ?= postgres://app:app@localhost:5432/content_digest?sslmode=disable
DOCKER_DATABASE_URL ?= postgres://app:app@postgres:5432/content_digest?sslmode=disable

.PHONY: test vet backend-run db-up db-down migrate-up migrate-down

test:
	cd backend && GOCACHE=/tmp/go-build go test ./...

vet:
	cd backend && GOCACHE=/tmp/go-build go vet ./...

backend-run:
	cd backend && go run ./cmd/server

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate-up:
	docker compose run --rm migrate -path=/migrations -database "$(DOCKER_DATABASE_URL)" up

migrate-down:
	docker compose run --rm migrate -path=/migrations -database "$(DOCKER_DATABASE_URL)" down
