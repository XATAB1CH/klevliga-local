# КлёвЛига

**КлёвЛига** — локальная кодовая база рыболовной социальной сети: тёмный community-интерфейс для форумов, отчётов о рыбалке, водоёмов, подписок и уведомлений.

Проект больше не является только frontend-starter: в репозитории есть рабочий React/Vite frontend и Go API с PostgreSQL, миграциями и частью бизнес-логики.

## Статус проекта

Реализовано:

- React/Vite/TypeScript frontend в `apps/web`.
- Go backend в `services/api`.
- REST API под `/api/v1`.
- Cookie-based авторизация через HttpOnly session cookie.
- PostgreSQL-схема и seed-данные через embedded Goose migrations.
- Форум, лента, водоёмы, отчёты, поиск, подписки, уведомления и модерационные жалобы на уровне MVP API.
- Локальный PostgreSQL через Docker Compose.

Пока не реализовано или реализовано частично:

- Загрузка фотографий в отчёты: общие media endpoints есть, но возвращают `NOT_IMPLEMENTED`. Аватары загружаются отдельным маршрутом.
- S3/object storage и очистка EXIF/GPS для изображений.
- Password recovery.
- PostGIS и точные геозапросы.
- Marketplace, AI assistant, realtime-функции — только как будущие/placeholder-направления.
- Полноценный список подписок и редактирование профиля.

## Технологии

### Frontend

- React 19
- Vite 6
- TypeScript
- React Router
- TanStack Query
- CSS, единая тёмная тема «Глубокая вода»

### Backend

- Go 1.26
- chi router
- pgx / pgxpool
- sqlc
- Goose migrations
- PostgreSQL 17
- Argon2id для паролей
- server-side sessions через HttpOnly cookie

## Структура проекта

```text
.
├── api/
│   └── openapi.yaml              # Черновик/контракт REST API
├── apps/
│   └── web/                      # React/Vite frontend
├── docs/                         # Архитектура, модель данных, заметки по переносу
├── infra/
│   └── docker-compose.yml        # Локальный PostgreSQL
└── services/
    └── api/                      # Go backend
```

Ключевые frontend-файлы:

- `apps/web/src/main.tsx` — вход React-приложения.
- `apps/web/src/App.tsx` — layout и маршруты.
- `apps/web/src/api/client.ts` — общий API-клиент.
- `apps/web/src/api/hooks.ts` — TanStack Query hooks/mutations.
- `apps/web/src/auth/AuthContext.tsx` — состояние авторизации.
- `apps/web/src/styles/global.css` — основной стиль интерфейса.

Ключевые backend-файлы:

- `services/api/cmd/klevliga/main.go` — вход приложения.
- `services/api/internal/app/app.go` — сборка приложения, БД, миграции, HTTP server.
- `services/api/internal/api/transport/http/server.go` — регистрация HTTP routes.
- `services/api/internal/services/` — бизнес-сервисы.
- `services/api/internal/storages/postgres/` — PostgreSQL repository/sqlc layer.
- `services/api/migrations/` — миграции БД.

## Быстрый запуск локально

Требования:

- Node.js LTS
- npm
- Go 1.26+
- PostgreSQL 17: напрямую либо через Docker и Docker Compose

Для Windows настроен запуск без Docker: [локальные инструменты, запуск и тесты](docs/LOCAL_DEVELOPMENT.md). Скрипты создают отдельную тестовую базу и хранят данные PostgreSQL вне OneDrive.

### 1. Запустить PostgreSQL

```bash
cd services/api
make db-up
```

База поднимается из `infra/docker-compose.yml`:

- host: `127.0.0.1`
- port: `5434`
- database: `klevliga`
- user: `klevliga`
- password по умолчанию: `klevliga_dev`

### 2. Настроить backend config

В `services/api/config.yaml` должен быть локальный конфиг для API. Пример находится в `services/api/config.example.yaml`.

Для локальной БД обычно нужны значения:

```yaml
server:
  http:
    address: ":8080"

connections:
  postgres:
    main:
      host: 127.0.0.1
      port: 5434
      user: klevliga
      password: klevliga_dev
      dbname: klevliga

auth:
  session_ttl: 720h
  cookie_name: klevliga_session
  cookie_secure: false
  cookie_domain: ""
  verification_ttl: 10m
  verification_secret: ${AUTH_VERIFICATION_SECRET:klevliga-local-development-only}

mail:
  mode: log

logger:
  title: klevliga-api
  level: debug
  pretty: true
```

### 3. Запустить backend

```bash
cd services/api
make run
```

API будет доступен на `http://localhost:8080/api/v1`.

При старте backend подключается к PostgreSQL и применяет embedded migrations.

### 4. Запустить frontend

В отдельном терминале:

```bash
cd apps/web
npm install
npm run dev
```

Vite обычно откроется на `http://localhost:5173`.

Frontend отправляет запросы на same-origin `/api/v1`; в dev-режиме Vite проксирует их на backend. Цель прокси можно переопределить через `apps/web/.env`:

```env
VITE_API_PROXY=http://127.0.0.1:8080
```

## Команды разработки

### Frontend

Из папки `apps/web`:

```bash
npm run dev        # dev server
npm run build      # TypeScript build + Vite build
npm run typecheck  # TypeScript check
npm run preview    # preview production build
npm run test       # Vitest: регрессии кэша и загрузки страниц
npm run test:e2e   # Playwright: UI, 393px и 1440px, клавиатура
```

### Backend

Из папки `services/api`:

```bash
make run              # запустить API
make build            # собрать бинарник
make test             # go mod tidy + unit tests с race detector
make test-unit        # unit tests
make test-integration # integration tests с build tag integration
make gen              # go generate + sqlc generate
make lint             # golangci-lint run
make db-up            # поднять PostgreSQL
make db-down          # остановить PostgreSQL
```

## API

Базовый путь API: `/api/v1`.

Актуальные особенности DTO, авторизации, ограничений и незавершённых функций описаны в [docs/API_NOTES.md](docs/API_NOTES.md). Для frontend используются Vitest/Testing Library и Playwright. Backend содержит unit tests и PostgreSQL integration tests с изолированными схемами (`-tags=integration`, переменная `KLEVLIGA_TEST_DATABASE_URL`).

Основные группы endpoints:

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/me`
- `GET /users/{id}`
- `GET /feed`
- `GET /forum/categories`
- `GET /forum/topics`
- `POST /forum/topics`
- `GET /forum/topics/{id}`
- `GET /forum/topics/{id}/posts`
- `POST /forum/topics/{id}/posts`
- `GET /reports`
- `POST /reports`
- `GET /reports/{id}`
- `GET /waters`
- `GET /waters/{id}`
- `GET /search`
- `POST /follows`
- `DELETE /follows/{target_type}/{target_id}`
- `GET /notifications`
- `POST /media/uploads`
- `POST /media/uploads/{id}/complete`
- `POST /moderation/flags`

Формат ошибок API:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "request_id": "..."
  }
}
```

## Данные и миграции

Основная миграция создаёт таблицы:

- `users`
- `user_profiles`
- `sessions`
- `regions`
- `forum_categories`
- `topics`
- `topic_posts`
- `waters`
- `reports`
- `follows`
- `notifications`
- `abuse_reports`

Seed-данные включают Московскую область, набор форумных категорий и несколько водоёмов.

## Принципы продукта и разработки

- Интерфейс должен оставаться тёмным, атмосферным, «рыболовным», а не превращаться в generic admin dashboard.
- Точные GPS-точки рыбалки приватны по умолчанию.
- EXIF/GPS из будущих пользовательских изображений нельзя публиковать наружу.
- Публичные фотографии водоёмов в текущем UI — иллюстративные, им нельзя приписывать реальные географические места без подтверждения.
- Новые backend endpoints должны идти под `/api/v1`.
- Секреты и реальные credentials не хранятся в репозитории.

## Проверка перед изменениями

Минимальная проверка frontend:

```bash
cd apps/web
npm run build
```

Минимальная проверка backend:

```bash
cd services/api
make test
make build
```

Для UI-изменений дополнительно проверять адаптивность примерно на ширинах `393px` и `1440px`, клавиатурную навигацию и понятность ошибок.

## Происхождение frontend

Первоначально frontend был создан как самостоятельный Vite starter по согласованному дизайну Floot. Это не официальный ZIP-экспорт Floot и не обход его системы экспорта. Сейчас проект развивается как самостоятельная локальная кодовая база с полноценным backend.
