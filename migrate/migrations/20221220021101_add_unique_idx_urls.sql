-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls
    ADD CONSTRAINT urls_unique_idx
        UNIQUE (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls
    DROP CONSTRAINT urls_unique_idx;
-- +goose StatementEnd
