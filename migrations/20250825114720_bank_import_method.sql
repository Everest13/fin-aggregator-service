-- +goose Up
CREATE TABLE IF NOT EXISTS bank_import_method
(
    bank_id INT NOT NULL,
    import_method varchar(20) NOT NULL,
    UNIQUE(bank_id, import_method)
);

-- +goose Down
DROP TABLE IF EXISTS bank_import_method;
