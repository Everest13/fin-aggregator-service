-- +goose Up
CREATE TABLE bank_header_mapping (
     header_id INT NOT NULL,
     transaction_field VARCHAR(50) NOT NULL,
     UNIQUE(header_id, transaction_field)
);

-- +goose Down
DROP TABLE IF EXISTS bank_header_mapping;
