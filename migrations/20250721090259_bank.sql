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
     required bool default false,
     UNIQUE(bank_id, name)
);

-- +goose Down
DROP TABLE IF EXISTS bank;
DROP TABLE IF EXISTS bank_header;
