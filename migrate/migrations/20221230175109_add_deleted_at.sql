-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE urls ADD COLUMN deleted_at timestamp;
-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN deleted_at;
-- +goose StatementEnd
