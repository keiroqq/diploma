# Content Digest App

PWA/MVP для агрегации и персонализации текстового контента из RSS/Atom-источников. Первый этап проекта сосредоточен на backend: пользователи, ленты, источники, правила фильтрации, RSS-обновление, выдача сегодняшних и архивных материалов, избранное.

## Стек

- Backend: Go, chi, GORM, PostgreSQL, SQL-миграции, JWT, bcrypt, gofeed, slog.
- Frontend: React, TypeScript, Vite, React Router, TanStack Query, Zustand, lucide-react.
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
make db-up
```

2. Создать `backend/.env` на основе `backend/.env.example`.

3. Применить миграции через контейнер `golang-migrate`:

```bash
make migrate-up
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
make swagger      # сгенерировать backend/docs для Swagger UI
make backend-run  # запуск backend API
make db-up        # PostgreSQL через docker compose
make migrate-up   # применить SQL-миграции через контейнер
make migrate-down # откатить SQL-миграции через контейнер
```

Swagger UI доступен после запуска backend:

```text
http://localhost:8080/swagger/index.html
```

## Запуск frontend

Первый MVP фронтенда находится в `frontend/` и по умолчанию обращается к API
`http://localhost:8080`.

```bash
cd frontend
npm install
npm run dev
```

Vite dev server открывает приложение на:

```text
http://localhost:5173
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
- `GET /api/catalog/topics`
- `POST /api/catalog/discover`
- `POST /api/feeds/{id}/catalog-sources`
- `GET /api/categories`
- `POST /api/sources/{id}/refresh`
- `POST /api/feeds/{id}/refresh`
- `GET /api/feeds/{id}/items?mode=today&limit=20`
- `GET /api/feeds/{id}/items?mode=archive&cursor=...&limit=20`
- `GET /api/feeds/{id}/items?mode=today&category=ai&limit=20`
- `POST /api/items/{id}/save`
- `DELETE /api/items/{id}/save`
- `GET /api/saved`
- `GET|POST /api/feeds/{id}/rules`
- `PUT|DELETE /api/feeds/{id}/rules/{ruleId}`

## Каталог тем

Backend хранит стартовый каталог тем приложения и страниц Хабра с новостями. RSS-ссылка не хардкодится: endpoint `POST /api/catalog/discover` скачивает HTML страницы и ищет `<link type="application/rss+xml">`.

Пример подключения тем к ленте:

```bash
curl -X POST http://localhost:8080/api/feeds/$FEED_ID/catalog-sources \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"source_ids":["habr-backend-news","habr-frontend-news","habr-ai-ml-news"]}'
```

## Нормализация тегов

RSS-источники возвращают сырые теги по-разному: например `ии`, `AI`, `ML`, `искусственный интеллект`, `нейросети`. Backend сохраняет исходные теги в `tags`, а также связывает материал с нормализованными категориями приложения через `categories`.

Пример категорий:

- `Backend`
- `Frontend`
- `DevOps`
- `Databases`
- `Security`
- `AI`
- `Design`
- `Management`
- `Hardware`

В ответе `/api/feeds/{id}/items` у материала есть два поля:

```json
{
  "tags": ["искусственный интеллект", "llm"],
  "categories": ["AI"]
}
```

Фильтр `target_type=tag` ищет и по исходным тегам, и по нормализованным категориям.

Категории можно получить отдельным endpoint:

```bash
curl http://localhost:8080/api/categories \
  -H "Authorization: Bearer $TOKEN"
```

Материалы можно фильтровать по slug категории:

```bash
curl "http://localhost:8080/api/feeds/$FEED_ID/items?mode=today&category=ai&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```
