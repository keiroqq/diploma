# Content Digest

Content Digest — PWA-приложение для агрегации, чтения и персонализации текстового контента из RSS/Atom-источников. Пользователь создает тематические потоки, подключает к ним источники из каталога или собственные RSS-ленты, обновляет материалы, фильтрует публикации по датам и темам, ищет по сохраненному корпусу статей и добавляет важные материалы в избранное.

Приложение ориентировано на мобильный сценарий: интерфейс адаптирован под смартфоны, поддерживает установку на главный экран, встроенную читалку статей и работу с несколькими потоками в формате новостной ленты.

## Возможности

- регистрация и вход по JWT;
- создание, редактирование и удаление пользовательских потоков;
- подключение RSS-источников к потокам;
- каталог готовых источников по тематическим разделам;
- поддержка источников Habr, Sports.ru, Ведомости и Коммерсантъ;
- ручное обновление одного источника или всего потока;
- сохранение материалов из RSS в PostgreSQL;
- нормализация тегов и категорий для фильтрации;
- фильтрация материалов по диапазону дат и темам;
- поиск по доступным пользователю материалам;
- избранные материалы;
- встроенная читалка с загрузкой полного текста статьи, когда это возможно;
- PWA manifest, service worker и иконки приложения;
- контейнерный запуск backend, frontend, PostgreSQL и миграций через Docker Compose.

## Стек

- Backend: Go, chi, GORM, PostgreSQL, SQL migrations, JWT, bcrypt, gofeed, goquery, bluemonday, slog.
- Frontend: React, TypeScript, Vite, React Router, TanStack Query, Zustand, lucide-react.
- Инфраструктура: Docker Compose, Nginx для раздачи frontend и proxy `/api` в backend.
- PWA: `vite-plugin-pwa`, Workbox, manifest, service worker.

## Структура проекта

```text
backend/
  cmd/server/             # точка входа backend API
  docs/                   # Swagger-документация
  internal/app/           # сборка приложения, роутер, зависимости
  internal/auth/          # регистрация, вход, JWT
  internal/catalog/       # каталог готовых RSS-источников
  internal/categories/    # нормализованные категории и категории потока
  internal/feeds/         # пользовательские потоки
  internal/fetch/         # безопасный HTTP-клиент для внешних источников
  internal/filters/       # правила фильтрации потоков
  internal/items/         # материалы, поиск, избранное, читалка
  internal/rss/           # загрузка и нормализация RSS/Atom
  migrations/             # SQL-миграции базы данных
  Dockerfile

frontend/
  public/                 # PWA-иконки и логотип
  src/api/                # API-клиент и DTO-типы
  src/components/         # общие UI-компоненты
  src/pages/              # страницы приложения
  src/store/              # клиентское состояние
  src/utils/              # форматирование, локальные материалы, настройки
  Dockerfile
  nginx.conf              # отдача SPA и proxy в backend

docker-compose.yml        # postgres, migrate, backend, frontend
docker-compose.prod.yml   # production override для запуска готовых GHCR-образов
.env.example              # шаблон переменных для Docker Compose
Makefile                  # команды локальной разработки и серверного запуска
```

## Быстрый запуск через Docker Compose

Скопировать шаблон переменных:

```bash
cp .env.example .env
```

Для локального запуска можно оставить значения по умолчанию. Для сервера обязательно заменить `JWT_SECRET`, `POSTGRES_PASSWORD` и соответствующий пароль внутри `DATABASE_URL`.

Запуск:

```bash
make local-up
```

Команда собирает backend и frontend из локальных исходников и запускает весь стек через `docker-compose.yml`.

При запуске выполняется цепочка:

1. стартует PostgreSQL;
2. Docker Compose ждет healthcheck базы;
3. контейнер `migrate` применяет SQL-миграции;
4. запускается backend;
5. после healthcheck backend запускается frontend/Nginx.

Приложение будет доступно по адресу:

```text
http://localhost:8080
```

Проверка сервисов:

```bash
make local-ps
make local-logs
curl http://localhost:8080/health
```

Остановка без удаления данных:

```bash
make local-down
```

Не используйте `docker compose down -v`, если нужно сохранить данные PostgreSQL.

## Локальный запуск для разработки

Поднять только PostgreSQL:

```bash
make db-up
```

Применить миграции:

```bash
make migrate-up
```

Запустить backend:

```bash
make backend-run
```

Запустить frontend:

```bash
make frontend-install
make frontend-dev
```

Frontend dev server доступен на:

```text
http://localhost:5173
```

В dev-режиме Vite проксирует `/api`, `/health` и `/swagger` на backend `http://localhost:8080`.

## Основные команды

Посмотреть список команд:

```bash
make help
```

Backend и frontend для разработки:

```bash
make test              # go test ./...
make vet               # go vet ./...
make swagger           # обновить Swagger docs
make backend-run       # запустить backend локально
make frontend-install  # установить frontend-зависимости
make frontend-dev      # запустить Vite dev server
make frontend-build    # собрать frontend
```

Локальный Docker Compose:

```bash
make local-config  # проверить итоговый compose-конфиг
make local-build   # собрать backend и frontend
make local-up      # собрать и запустить весь стек
make local-down    # остановить стек
make local-logs    # смотреть логи
make local-ps      # показать сервисы
```

База и миграции для локальной разработки:

```bash
make db-up         # поднять PostgreSQL
make db-down       # остановить compose-сервисы
make migrate-up    # применить миграции
make migrate-down  # откатить миграции
```

Серверный запуск из GitHub Container Registry:

```bash
make server-config   # проверить prod-конфиг
make server-pull     # скачать backend/frontend образы
make server-up       # запустить стек из готовых образов
make server-deploy   # скачать свежие образы и запустить стек
make server-down     # остановить стек
make server-logs     # смотреть логи
make server-ps       # показать сервисы
make server-restart  # перезапустить сервисы
make server-migrate  # вручную запустить контейнер миграций
```

## Переменные окружения

Корневой `.env` используется Docker Compose.

```env
POSTGRES_USER=app
POSTGRES_PASSWORD=app
POSTGRES_DB=content_digest
POSTGRES_PORT=5432

APP_ENV=production
HTTP_PORT=8080
DATABASE_URL=postgres://app:app@postgres:5432/content_digest?sslmode=disable
JWT_SECRET=change_me_replace_before_deploy
JWT_EXPIRES_IN=24h
LOG_LEVEL=info
RSS_REFRESH_COOLDOWN=15m

FRONTEND_PORT=8080
VITE_API_URL=
```

`VITE_API_URL=` оставлен пустым намеренно: в production frontend делает same-origin запросы вида `/api/auth/login`, а Nginx проксирует `/api` в backend.

## API

Все endpoint'ы внутри `/api`, кроме регистрации и входа, требуют заголовок:

```text
Authorization: Bearer <jwt>
```

Служебные endpoint'ы:

- `GET /health` — healthcheck backend.
- `GET /swagger/index.html` — Swagger UI.

### Auth

- `POST /api/auth/register` — регистрация пользователя.
- `POST /api/auth/login` — вход и получение JWT.
- `GET /api/auth/me` — текущий пользователь.

### Потоки

- `GET /api/feeds` — список потоков пользователя.
- `POST /api/feeds` — создать поток.
- `GET /api/feeds/{id}` — получить поток.
- `PUT /api/feeds/{id}` — обновить поток.
- `DELETE /api/feeds/{id}` — удалить поток.
- `GET /api/feeds/{id}/sources` — источники потока.
- `POST /api/feeds/{id}/sources` — подключить источник к потоку.
- `DELETE /api/feeds/{id}/sources/{sourceId}` — отключить источник от потока.
- `POST /api/feeds/{id}/refresh` — обновить все источники потока.

### Источники

- `GET /api/sources` — доступные пользователю источники.
- `POST /api/sources` — создать RSS-источник.
- `GET /api/sources/{id}` — получить источник.
- `PUT /api/sources/{id}` — обновить источник.
- `DELETE /api/sources/{id}` — удалить пользовательский источник.
- `POST /api/sources/{id}/refresh` — обновить источник.
- `GET /api/sources/{id}/preview-items` — предварительный просмотр материалов источника.

### Каталог

- `GET /api/catalog/topics` — тематический каталог готовых источников.
- `POST /api/catalog/discover` — найти RSS на странице сайта.
- `POST /api/feeds/{id}/catalog-sources` — подключить выбранные источники каталога к потоку.

Каталог сгруппирован по разделам: IT и технологии, бизнес и экономика, политика и общество, происшествия, наука и железо, спорт, недвижимость и авто, культура и стиль жизни, регионы.

### Категории

- `GET /api/categories` — все нормализованные категории приложения.
- `GET /api/feeds/{id}/categories` — категории, доступные в конкретном потоке.

### Материалы

- `GET /api/feeds/{id}/items` — материалы потока.
- `GET /api/items/search` — поиск по материалам.
- `GET /api/items/{id}` — материал для встроенной читалки.
- `POST /api/items/{id}/save` — добавить материал в избранное.
- `DELETE /api/items/{id}/save` — удалить материал из избранного.
- `GET /api/saved` — список избранных материалов.

Основные query-параметры для `/api/feeds/{id}/items`:

- `mode=today|archive|all`;
- `limit=20`;
- `cursor=<RFC3339>` для архивной пагинации;
- `category=ai`;
- `categories=ai,backend`;
- `date_from=YYYY-MM-DD`;
- `date_to=YYYY-MM-DD`.

Пример:

```bash
curl "http://localhost:8080/api/feeds/$FEED_ID/items?mode=all&categories=ai,backend&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

Поиск:

```bash
curl "http://localhost:8080/api/items/search?q=postgres&feed_id=$FEED_ID&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

### Правила фильтрации

- `GET /api/feeds/{id}/rules` — правила фильтрации потока.
- `POST /api/feeds/{id}/rules` — создать правило.
- `PUT /api/feeds/{id}/rules/{ruleId}` — обновить правило.
- `DELETE /api/feeds/{id}/rules/{ruleId}` — удалить правило.

## PWA и деплой

Frontend собирается как PWA: в production build попадают `manifest.webmanifest`, `sw.js`, Workbox-файлы и иконки приложения.

Для установки приложения на главный экран телефона нужен HTTPS-домен. Типовая production-схема:

```text
https://example.com/      -> frontend/Nginx
https://example.com/api/* -> backend через Nginx proxy
```

Текущий Docker Compose публикует frontend на `FRONTEND_PORT`. Для реального домена можно поставить внешний reverse proxy, например Caddy или Nginx, и направить HTTPS-трафик на этот порт.

### Запуск на сервере из GHCR

GitHub Actions публикует Docker-образы:

```text
ghcr.io/keiroqq/diploma-backend:latest
ghcr.io/keiroqq/diploma-frontend:latest
```

На сервере нужен `.env` на основе `.env.example` и файлы compose-проекта. Если GHCR-пакеты закрыты, сначала выполните login:

```bash
echo <TOKEN> | docker login ghcr.io -u keiroqq --password-stdin
```

Запуск production-конфигурации:

```bash
make server-deploy
```

Эта команда использует два compose-файла:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml ...
```

`docker-compose.yml` описывает общую инфраструктуру: PostgreSQL, миграции, переменные, healthchecks и порты. `docker-compose.prod.yml` переопределяет backend, frontend и migrator: вместо локальной сборки и bind mount миграций используются готовые образы из GHCR.

Backend-образ содержит миграции и CLI для их применения:

```text
/app/migrations
/app/migrate
```

В production compose сервис `migrate` использует тот же backend-образ, поэтому серверу или Kubernetes Job не нужен отдельный bind mount папки `backend/migrations`. Для Kubernetes миграции можно запускать командой:

```bash
/app/migrate -path=/app/migrations -database "$DATABASE_URL" up
```

Для запуска конкретного тега:

```bash
make server-deploy IMAGE_TAG=main
make server-deploy IMAGE_TAG=sha-<commit>
```

## Лицензия

Проект разработан в рамках дипломной работы.
