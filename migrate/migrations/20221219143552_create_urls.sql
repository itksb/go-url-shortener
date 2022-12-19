-- +goose Up
-- +goose StatementBegin
CREATE TABlE IF NOT EXISTS urls
(
    id           SERIAL PRIMARY KEY,
    user_id      CHARACTER VARYING(36) NOT NULL,
    original_url CHARACTER VARYING     NOT NULL,

    created_at   TIMESTAMP             NOT NULL DEFAULT now(),
    updated_at   TIMESTAMP             NOT NULL DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
