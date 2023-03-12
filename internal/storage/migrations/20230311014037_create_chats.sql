-- +goose Up
-- +goose StatementBegin
CREATE TABLE chats
(
    id          BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT    NOT NULL UNIQUE,
    topics      TEXT[] NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- drop table if exists:
DROP TABLE IF EXISTS chats;
-- +goose StatementEnd
