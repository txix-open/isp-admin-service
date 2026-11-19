-- +goose Up
ALTER TABLE roles ADD COLUMN application_group TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE roles DROP COLUMN application_group;
