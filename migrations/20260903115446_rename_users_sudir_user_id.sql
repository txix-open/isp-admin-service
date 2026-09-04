-- +goose Up
ALTER TABLE users RENAME COLUMN sudir_user_id TO oauth_user_id;

-- +goose Down
ALTER TABLE users RENAME COLUMN oauth_user_id TO sudir_user_id;
