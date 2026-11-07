-- +goose Up
ALTER TABLE user_profiles ADD COLUMN roles text[] NOT NULL DEFAULT ARRAY['user']::text[];
UPDATE user_profiles SET roles = CASE WHEN role = 'user' THEN ARRAY['user'] ELSE ARRAY['user', role] END;
ALTER TABLE user_profiles DROP COLUMN role;
ALTER TABLE user_profiles ADD CONSTRAINT user_profiles_roles_check CHECK (
    roles @> ARRAY['user']::text[]
    AND roles <@ ARRAY['user', 'moderator', 'support', 'admin']::text[]
    AND cardinality(roles) BETWEEN 1 AND 4
    AND array_position(roles, NULL) IS NULL
    AND cardinality(roles) = (
        CASE WHEN 'user' = ANY(roles) THEN 1 ELSE 0 END +
        CASE WHEN 'moderator' = ANY(roles) THEN 1 ELSE 0 END +
        CASE WHEN 'support' = ANY(roles) THEN 1 ELSE 0 END +
        CASE WHEN 'admin' = ANY(roles) THEN 1 ELSE 0 END
    )
);

-- Serialize role/status changes, including first-admin bootstrap. The lock is
-- local to this schema and survives neither transaction rollback nor commit.
CREATE TABLE access_control_lock (id boolean PRIMARY KEY CHECK (id));
INSERT INTO access_control_lock VALUES (true);

CREATE TABLE audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type text NOT NULL CHECK (actor_type IN ('user', 'bootstrap')),
    actor_id uuid REFERENCES users(id),
    action text NOT NULL,
    target_type text NOT NULL,
    target_id uuid NOT NULL,
    reason text NOT NULL CHECK (char_length(reason) BETWEEN 3 AND 1000),
    request_id text NOT NULL CHECK (char_length(request_id) BETWEEN 1 AND 100),
    changes jsonb NOT NULL CHECK (jsonb_typeof(changes) = 'object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT audit_actor_check CHECK (
        (actor_type = 'user' AND actor_id IS NOT NULL)
        OR (actor_type = 'bootstrap' AND actor_id IS NULL AND action = 'admin.bootstrap')
    )
);
CREATE INDEX audit_events_time_idx ON audit_events (created_at DESC, id DESC);
CREATE INDEX audit_events_actor_idx ON audit_events (actor_id, created_at DESC, id DESC);
CREATE INDEX audit_events_target_idx ON audit_events (target_id, created_at DESC, id DESC);

-- +goose StatementBegin
CREATE FUNCTION reject_audit_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only';
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER audit_events_append_only BEFORE UPDATE OR DELETE OR TRUNCATE ON audit_events
FOR EACH STATEMENT EXECUTE FUNCTION reject_audit_mutation();

-- +goose Down
DROP TABLE audit_events;
DROP FUNCTION reject_audit_mutation();
DROP TABLE access_control_lock;
ALTER TABLE user_profiles ADD COLUMN role text NOT NULL DEFAULT 'user';
UPDATE user_profiles SET role = CASE
    WHEN roles @> ARRAY['admin']::text[] THEN 'admin'
    WHEN roles @> ARRAY['moderator']::text[] THEN 'moderator'
    ELSE 'user' END;
ALTER TABLE user_profiles ADD CONSTRAINT user_profiles_role_check CHECK (role IN ('user', 'moderator', 'admin'));
ALTER TABLE user_profiles DROP COLUMN roles;
