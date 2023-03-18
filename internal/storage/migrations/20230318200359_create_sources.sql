-- +goose Up
-- +goose StatementBegin
CREATE TABLE sources
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    feed_url   VARCHAR(255) NOT NULL,
    priority   INT          NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sources;
-- +goose StatementEnd
