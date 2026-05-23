DATABASE_URL ?= postgres://app:app@localhost:5432/content_digest?sslmode=disable
DOCKER_DATABASE_URL ?= postgres://app:app@postgres:5432/content_digest?sslmode=disable

.PHONY: test vet swagger backend-run db-up db-down migrate-up migrate-down

test:
	cd backend && GOCACHE=/tmp/go-build go test ./...

vet:
	cd backend && GOCACHE=/tmp/go-build go vet ./...

swagger:
	cd backend && GOCACHE=/tmp/go-build go run github.com/swaggo/swag/cmd/swag init --dir ./cmd/server,./internal/auth,./internal/catalog,./internal/categories,./internal/feeds,./internal/filters,./internal/items,./internal/sources,./internal/rss -g main.go -o docs --parseInternal

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
