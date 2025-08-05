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

DO $$
    DECLARE
        month_start DATE;
        month_end DATE;
        partition_name TEXT;
        i INT;
    BEGIN
        FOR i IN 0..11 LOOP
                month_start := date_trunc('year', CURRENT_DATE) + (i || ' month')::interval;
                month_end := month_start + interval '1 month';
                partition_name := format('transaction_%s', to_char(month_start, 'YYYY_MM'));

                EXECUTE format(
                        'CREATE TABLE IF NOT EXISTS %I PARTITION OF transaction
                         FOR VALUES FROM (%L) TO (%L);',
                        partition_name,
                        month_start,
                        month_end
                    );
            END LOOP;
    END $$;

-- +goose Down
DROP TABLE IF EXISTS transaction;
