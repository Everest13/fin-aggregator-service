-- +goose Up
CREATE TABLE IF NOT EXISTS transaction (
    id SERIAL,
    bank_id int NOT NULL,
    external_id VARCHAR(100),
    user_id INTEGER NOT NULL,
    transaction_date DATE NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    category_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp,
    type VARCHAR(20) NOT NULL,
    import_method VARCHAR(50) NOT NULL,
    PRIMARY KEY (id, transaction_date)
) PARTITION BY RANGE (transaction_date);

ALTER TABLE transaction
    ADD CONSTRAINT uniq_transaction_external
        UNIQUE (user_id, bank_id, external_id, amount, transaction_date);

-- +goose Down
DROP TABLE IF EXISTS transaction;
