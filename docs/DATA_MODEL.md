# Концепция схемы PostgreSQL (без реализации миграций)

Этот файл сохраняет первоначальную концепцию и не является точным описанием текущих таблиц. Фактическая схема находится в `services/api/migrations`; новые роли, чат, поддержка и аудит запланированы в [ROADMAP.md](ROADMAP.md). Перед их реализацией модель и контракты нужно актуализировать вместе с миграциями.

## Сущности

- `users (id UUID PK, email CITEXT UNIQUE, password_hash, username CITEXT UNIQUE, status, created_at, updated_at)`
- `user_profiles (user_id PK/FK, display_name, bio, region_id, avatar_media_id, role, updated_at)`
- `sessions (id UUID PK, user_id FK, token_hash, expires_at, revoked_at, created_at)` — серверные сессии/refresh rotation.
- `regions (id, name, slug)`
- `forum_categories (id, parent_id FK nullable, slug UNIQUE, name, position, status)`
- `topics (id, category_id FK, author_id FK, slug UNIQUE, title, body, type, pinned, locked, last_activity_at, created_at, updated_at, deleted_at)`
- `topic_posts (id, topic_id FK, author_id FK, body, parent_post_id FK nullable, created_at, edited_at, deleted_at)`
- `waters (id, slug UNIQUE, name, region_id FK, kind, description, point GEOGRAPHY(POINT,4326) nullable, visibility, created_at)`
- `reports (id, author_id FK, topic_id FK UNIQUE, water_id FK nullable, fished_at, species_ids[], location_visibility, details, created_at)` — отчёт связан с темой, поэтому у отчёта одна система комментариев `topic_posts`.
- `media (id, uploader_id FK, key, mime, size_bytes, width, height, status, public_url, created_at)`; `report_media (report_id, media_id, position)`; `topic_media` аналогично.
- `follows (user_id, target_type ENUM('topic','water'), target_id UUID, created_at, UNIQUE(user_id,target_type,target_id))`
- `notifications (id, recipient_id FK, type, actor_id FK, target_type, target_id, read_at, created_at)`
- `reports_abuse (id, reporter_id FK, target_type, target_id, reason, status, created_at)` — жалобы на нарушения, не путать с рыболовными отчётами.
- `moderation_actions (id, actor_id FK, target_type, target_id, action, reason, created_at)`.

## Индексы и технические решения

- `topics (category_id,last_activity_at DESC,id DESC)` для списка тем; `topic_posts (topic_id, created_at, id)` для сообщений.
- `reports (water_id,fished_at DESC,id DESC)` для карточки водоёма.
- `notifications (recipient_id,read_at,created_at DESC)`.
- PostgreSQL `GIN` на `to_tsvector('russian', title || ' ' || body)` для первичного поиска по форуму.
- PostGIS `GIST` на `waters.point` для будущих геозапросов (если геоданные внедряются в MVP — сразу; иначе отдельной миграцией).
- `follows.target_type/target_id` с полиморфным указателем удобны, но не дают простой FK; для строгой целостности лучше `topic_follows` и `water_follows` как отдельные таблицы. Зафиксировать решение до миграций.
- Soft-delete только для пользовательского контента, в модерации хранить причину и автора операции.
- Не хранить секретные точные координаты в общедоступном поле; отдельная таблица приватных точек/политика доступа.

## Важная предметная модель

**Один отчёт = структурированные данные + одна тема форума.** Отчёт отображается в общей ленте, карточке водоёма и профиле, но обсуждение/ответы не дублируются. Публичная точка по умолчанию — **водоём, не GPS**. Все media attachments имеют состояния `pending`, `uploaded`, `processing`, `ready`, `rejected`.
