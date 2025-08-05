-- +goose Up
CREATE TABLE IF NOT EXISTS user_bank
(
    user_id INT NOT NULL,
    bank_id INT NOT NULL,
    UNIQUE (user_id, bank_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_bank;
