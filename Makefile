DATABASE_URL ?= postgres://app:app@localhost:5432/content_digest?sslmode=disable
DOCKER_DATABASE_URL ?= postgres://app:app@postgres:5432/content_digest?sslmode=disable
COMPOSE ?= docker compose
COMPOSE_LOCAL ?= $(COMPOSE) -f docker-compose.yml
COMPOSE_SERVER ?= $(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml
IMAGE_TAG ?= latest
BACKEND_IMAGE ?= ghcr.io/keiroqq/diploma-backend
FRONTEND_IMAGE ?= ghcr.io/keiroqq/diploma-frontend
SERVER_ENV := IMAGE_TAG=$(IMAGE_TAG) BACKEND_IMAGE=$(BACKEND_IMAGE) FRONTEND_IMAGE=$(FRONTEND_IMAGE)

.PHONY: help test vet swagger backend-run frontend-install frontend-dev frontend-build db-up db-down migrate-up migrate-down local-config local-build local-up local-down local-logs local-ps server-config server-pull server-up server-deploy server-down server-logs server-ps server-restart server-migrate

help:
	@printf "%s\n" \
		"Backend/dev:" \
		"  make backend-run       run backend locally" \
		"  make test              run backend tests" \
		"  make vet               run go vet" \
		"  make swagger           regenerate Swagger docs" \
		"" \
		"Frontend/dev:" \
		"  make frontend-install  install frontend dependencies" \
		"  make frontend-dev      run Vite dev server" \
		"  make frontend-build    build frontend" \
		"" \
		"Local Docker:" \
		"  make local-up          build and start local compose stack" \
		"  make local-down        stop local compose stack" \
		"  make local-logs        follow local compose logs" \
		"  make local-ps          show local compose services" \
		"" \
		"Server Docker:" \
		"  make server-deploy     pull GHCR images and start stack" \
		"  make server-pull       pull GHCR images" \
		"  make server-up         start stack from GHCR images" \
		"  make server-logs       follow server compose logs"

test:
	cd backend && GOCACHE=/tmp/go-build go test ./...

vet:
	cd backend && GOCACHE=/tmp/go-build go vet ./...

swagger:
	cd backend && GOCACHE=/tmp/go-build go run github.com/swaggo/swag/cmd/swag init --dir ./cmd/server,./internal/auth,./internal/catalog,./internal/categories,./internal/feeds,./internal/filters,./internal/items,./internal/sources,./internal/rss -g main.go -o docs --parseInternal

backend-run:
	cd backend && go run ./cmd/server

frontend-install:
	cd frontend && npm ci

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

db-up:
	$(COMPOSE_LOCAL) up -d postgres

db-down:
	$(COMPOSE_LOCAL) down

migrate-up:
	$(COMPOSE_LOCAL) run --rm migrate -path=/migrations -database "$(DOCKER_DATABASE_URL)" up

migrate-down:
	$(COMPOSE_LOCAL) run --rm migrate -path=/migrations -database "$(DOCKER_DATABASE_URL)" down

local-config:
	$(COMPOSE_LOCAL) config

local-build:
	$(COMPOSE_LOCAL) build

local-up:
	$(COMPOSE_LOCAL) up -d --build

local-down:
	$(COMPOSE_LOCAL) down

local-logs:
	$(COMPOSE_LOCAL) logs -f

local-ps:
	$(COMPOSE_LOCAL) ps

server-config:
	$(SERVER_ENV) $(COMPOSE_SERVER) config

server-pull:
	$(SERVER_ENV) $(COMPOSE_SERVER) pull

server-up:
	$(SERVER_ENV) $(COMPOSE_SERVER) up -d

server-deploy:
	$(SERVER_ENV) $(COMPOSE_SERVER) pull
	$(SERVER_ENV) $(COMPOSE_SERVER) up -d

server-down:
	$(SERVER_ENV) $(COMPOSE_SERVER) down

server-logs:
	$(SERVER_ENV) $(COMPOSE_SERVER) logs -f

server-ps:
	$(SERVER_ENV) $(COMPOSE_SERVER) ps

server-restart:
	$(SERVER_ENV) $(COMPOSE_SERVER) restart

server-migrate:
	$(SERVER_ENV) $(COMPOSE_SERVER) run --rm migrate
