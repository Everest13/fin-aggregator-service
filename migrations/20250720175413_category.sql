-- +goose Up
CREATE TABLE IF NOT EXISTS category (
    id SERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    description text
);

CREATE TABLE category_keyword (
    id SERIAL PRIMARY KEY,
    category_id INTEGER NOT NULL,
    name VARCHAR(30) NOT NULL,
    UNIQUE (category_id, name)
);

-- +goose Down
DROP TABLE IF EXISTS category;
DROP TABLE IF EXISTS category_keyword;

