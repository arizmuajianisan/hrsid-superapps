-- +goose Up
-- +goose StatementBegin
ALTER TABLE applications ADD COLUMN sso_secret TEXT NOT NULL DEFAULT '';

CREATE TABLE sso_launch_tokens (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sso_launch_tokens_token ON sso_launch_tokens(token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sso_launch_tokens;
ALTER TABLE applications DROP COLUMN sso_secret;
-- +goose StatementEnd
