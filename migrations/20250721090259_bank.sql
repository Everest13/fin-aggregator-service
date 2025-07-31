-- +goose Up
CREATE TABLE IF NOT EXISTS bank
(
    id   SERIAL PRIMARY KEY,
    name varchar(20) NOT NULL
);

CREATE TABLE bank_header (
     id SERIAL PRIMARY KEY,
     bank_id INT  NOT NULL,
     name VARCHAR(50) NOT NULL,
     target_field TEXT[] NOT NULL,
     UNIQUE(bank_id, name, target_field)
);

-- +goose Down
DROP TABLE IF EXISTS bank;
DROP TABLE IF EXISTS bank_header;
