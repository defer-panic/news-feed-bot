-- +goose Up
-- +goose StatementBegin
ALTER TABLE articles ADD COLUMN summary TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE articles DROP COLUMN summary;
-- +goose StatementEnd
