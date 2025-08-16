-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id   SERIAL PRIMARY KEY,
    name varchar(50) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS users;
