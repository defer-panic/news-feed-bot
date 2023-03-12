-- +goose Up
-- +goose StatementBegin
CREATE TABLE articles
(
    id           SERIAL PRIMARY KEY,
    title        TEXT      NOT NULL,
    link         TEXT      NOT NULL UNIQUE,
    topic        TEXT      NOT NULL,
    topic_score  INTEGER   NOT NULL,
    source_name  TEXT      NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
