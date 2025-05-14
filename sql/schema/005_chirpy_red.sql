-- +goose Up
ALTER TABLE users add is_chirpy_red boolean DEFAULT false;




-- +goose Down 
ALTER TABLE users drop column is_chirpy_red;
