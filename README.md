# Content Digest App

PWA/MVP для агрегации и персонализации текстового контента из RSS/Atom-источников. Первый этап проекта сосредоточен на backend: пользователи, ленты, источники, правила фильтрации, RSS-обновление, выдача сегодняшних и архивных материалов, избранное.

## Стек

- Backend: Go, chi, GORM, PostgreSQL, SQL-миграции, JWT, bcrypt, gofeed, slog.
- Frontend: React, TypeScript, Vite, TanStack Query, Zustand, PWA. Будет добавлен после стабилизации backend.
- Redis в MVP не используется.

## Структура

```text
backend/
  cmd/server          # точка входа API
  internal/auth       # регистрация, вход, JWT
  internal/feeds      # пользовательские ленты
  internal/sources    # RSS-источники
  internal/items      # выдача материалов и избранное
  internal/filters    # правила фильтрации
  internal/rss        # загрузка, очистка и нормализация RSS
  migrations          # SQL-схема для golang-migrate
frontend/             # место под PWA
docker-compose.yml
```

## Запуск backend

1. Запустить PostgreSQL:

```bash
docker compose up -d
```

2. Создать `backend/.env` на основе `backend/.env.example`.

3. Применить миграции через `golang-migrate`:

```bash
migrate -path backend/migrations -database "postgres://app:app@localhost:5432/content_digest?sslmode=disable" up
```

4. Запустить API:

```bash
cd backend
go run ./cmd/server
```

Healthcheck: `GET http://localhost:8080/health`.

## Dev-команды

```bash
make test         # go test ./...
make vet          # go vet ./...
make backend-run  # запуск backend API
make db-up        # PostgreSQL через docker compose
make migrate-up   # применить SQL-миграции
```

## Основные endpoint'ы MVP

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me`
- `GET|POST /api/feeds`
- `GET|PUT|DELETE /api/feeds/{id}`
- `POST /api/feeds/{id}/sources`
- `DELETE /api/feeds/{id}/sources/{sourceId}`
- `GET|POST /api/sources`
- `GET|PUT|DELETE /api/sources/{id}`
- `POST /api/sources/{id}/refresh`
- `POST /api/feeds/{id}/refresh`
- `GET /api/feeds/{id}/items?mode=today&limit=20`
- `GET /api/feeds/{id}/items?mode=archive&cursor=...&limit=20`
- `POST /api/items/{id}/save`
- `DELETE /api/items/{id}/save`
- `GET /api/saved`
- `GET|POST /api/feeds/{id}/rules`
- `PUT|DELETE /api/feeds/{id}/rules/{ruleId}`
