-- +goose Up
ALTER TABLE users add hashed_password text DEFAULT 'unset';



-- +goose Down 
ALTER TABLE users drop column hashed_password;
