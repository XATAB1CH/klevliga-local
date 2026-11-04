-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         citext NOT NULL UNIQUE,
    username      citext NOT NULL UNIQUE,
    password_hash text NOT NULL,
    status        text NOT NULL DEFAULT 'active',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'blocked'))
);

CREATE TABLE user_profiles (
    user_id      uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    display_name text NOT NULL,
    bio          text,
    region       text,
    avatar_url   text,
    role         text NOT NULL DEFAULT 'user',
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_profiles_role_check CHECK (role IN ('user', 'moderator', 'admin'))
);

CREATE TABLE sessions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);

CREATE TABLE regions (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    slug text NOT NULL UNIQUE
);

CREATE TABLE forum_categories (
    id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id uuid REFERENCES forum_categories (id),
    slug      text NOT NULL UNIQUE,
    name      text NOT NULL,
    position  integer NOT NULL DEFAULT 0,
    status    text NOT NULL DEFAULT 'active',
    CONSTRAINT forum_categories_status_check CHECK (status IN ('active', 'hidden'))
);

CREATE TABLE topics (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id      uuid NOT NULL REFERENCES forum_categories (id),
    author_id        uuid NOT NULL REFERENCES users (id),
    slug             text NOT NULL UNIQUE,
    title            text NOT NULL,
    body             text NOT NULL,
    type             text NOT NULL,
    pinned           boolean NOT NULL DEFAULT false,
    locked           boolean NOT NULL DEFAULT false,
    replies_count    integer NOT NULL DEFAULT 0,
    views_count      integer NOT NULL DEFAULT 0,
    last_activity_at timestamptz NOT NULL DEFAULT now(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    CONSTRAINT topics_type_check CHECK (type IN ('discussion', 'question', 'report'))
);

CREATE INDEX topics_feed_idx ON topics (last_activity_at DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX topics_category_idx ON topics (category_id, last_activity_at DESC, id DESC) WHERE deleted_at IS NULL;

CREATE TABLE topic_posts (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id       uuid NOT NULL REFERENCES topics (id) ON DELETE CASCADE,
    author_id      uuid NOT NULL REFERENCES users (id),
    body           text NOT NULL,
    parent_post_id uuid REFERENCES topic_posts (id),
    created_at     timestamptz NOT NULL DEFAULT now(),
    edited_at      timestamptz,
    deleted_at     timestamptz
);

CREATE INDEX topic_posts_topic_idx ON topic_posts (topic_id, created_at, id) WHERE deleted_at IS NULL;

CREATE TABLE waters (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        text NOT NULL UNIQUE,
    name        text NOT NULL,
    region_id   uuid REFERENCES regions (id),
    kind        text NOT NULL,
    description text,
    visibility  text NOT NULL DEFAULT 'public',
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT waters_kind_check CHECK (kind IN ('river', 'lake', 'reservoir', 'pond', 'other')),
    CONSTRAINT waters_visibility_check CHECK (visibility IN ('public', 'hidden'))
);

CREATE TABLE reports (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id           uuid NOT NULL REFERENCES users (id),
    topic_id            uuid NOT NULL UNIQUE REFERENCES topics (id),
    water_id            uuid REFERENCES waters (id),
    title               text NOT NULL,
    body                text NOT NULL,
    fished_at           date NOT NULL,
    location_visibility text NOT NULL DEFAULT 'water_only',
    species             text[] NOT NULL DEFAULT '{}',
    created_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT reports_visibility_check CHECK (location_visibility IN ('water_only', 'private'))
);

CREATE INDEX reports_water_idx ON reports (water_id, fished_at DESC, id DESC);

CREATE TABLE follows (
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    target_type text NOT NULL,
    target_id   uuid NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, target_type, target_id),
    CONSTRAINT follows_target_type_check CHECK (target_type IN ('topic', 'water'))
);

CREATE TABLE notifications (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type         text NOT NULL,
    actor_id     uuid REFERENCES users (id),
    target_type  text,
    target_id    uuid,
    title        text NOT NULL,
    target_url   text,
    read_at      timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX notifications_recipient_idx ON notifications (recipient_id, created_at DESC, id DESC);

CREATE TABLE abuse_reports (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id uuid NOT NULL REFERENCES users (id),
    target_type text NOT NULL,
    target_id   uuid NOT NULL,
    reason      text NOT NULL,
    status      text NOT NULL DEFAULT 'open',
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT abuse_reports_target_check CHECK (target_type IN ('topic', 'post', 'report', 'user')),
    CONSTRAINT abuse_reports_status_check CHECK (status IN ('open', 'resolved', 'rejected'))
);

INSERT INTO regions (name, slug) VALUES
    ('Московская область', 'moskovskaya-oblast');

INSERT INTO forum_categories (slug, name, position) VALUES
    ('general', 'Общие вопросы', 10),
    ('reports', 'Отчёты о рыбалке', 20),
    ('waters', 'Водоёмы и места', 30),
    ('spinning', 'Спиннинг', 40),
    ('feeder', 'Фидер', 50),
    ('float', 'Поплавочная ловля', 60),
    ('ice', 'Зимняя рыбалка', 70),
    ('tackle', 'Снасти и снаряжение', 80),
    ('boats', 'Лодки и техника', 90),
    ('tourism', 'Туризм и отдых', 100),
    ('cooking', 'Готовим улов', 110),
    ('other', 'Разное', 120);

INSERT INTO waters (slug, name, region_id, kind, description)
SELECT v.slug, v.name, r.id, v.kind, v.description
FROM (VALUES
    ('moskva', 'Москва-река', 'river', 'Иллюстративная карточка. Точные точки уловов не публикуются.'),
    ('oka', 'Ока', 'river', 'Иллюстративная карточка. Точные точки уловов не публикуются.'),
    ('istra', 'Истринское водохранилище', 'reservoir', 'Иллюстративная карточка. Точные точки уловов не публикуются.'),
    ('ruza', 'Рузское водохранилище', 'reservoir', 'Иллюстративная карточка. Точные точки уловов не публикуются.')
) AS v(slug, name, kind, description)
CROSS JOIN regions r
WHERE r.slug = 'moskovskaya-oblast';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS abuse_reports;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS waters;
DROP TABLE IF EXISTS topic_posts;
DROP TABLE IF EXISTS topics;
DROP TABLE IF EXISTS forum_categories;
DROP TABLE IF EXISTS regions;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
