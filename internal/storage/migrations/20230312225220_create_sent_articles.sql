-- +goose Up
-- +goose StatementBegin
CREATE TABLE sent_articles
(
    id               SERIAL PRIMARY KEY,
    article_id       INT       NOT NULL,
    chat_telegram_id INT       NOT NULL,
    sent_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,
    FOREIGN KEY (chat_telegram_id) REFERENCES chats (telegram_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX sent_articles_article_id_chat_telegram_id_uindex ON sent_articles (article_id, chat_telegram_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sent_articles;

DROP INDEX IF EXISTS sent_articles_article_id_chat_telegram_id_uindex;
-- +goose StatementEnd
