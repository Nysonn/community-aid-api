-- +goose Up
ALTER TABLE emergency_requests
    ADD COLUMN target_amount            NUMERIC(12,2),
    ADD COLUMN payment_type             TEXT CHECK (payment_type IN ('bank', 'mobile_money')),
    ADD COLUMN bank_account_name        TEXT,
    ADD COLUMN bank_account_number      TEXT,
    ADD COLUMN bank_name                TEXT,
    ADD COLUMN receiving_mobile_provider TEXT CHECK (receiving_mobile_provider IN ('mtn_momo', 'airtel_money')),
    ADD COLUMN receiving_mobile_number  TEXT;

-- +goose Down
ALTER TABLE emergency_requests
    DROP COLUMN IF EXISTS receiving_mobile_number,
    DROP COLUMN IF EXISTS receiving_mobile_provider,
    DROP COLUMN IF EXISTS bank_name,
    DROP COLUMN IF EXISTS bank_account_number,
    DROP COLUMN IF EXISTS bank_account_name,
    DROP COLUMN IF EXISTS payment_type,
    DROP COLUMN IF EXISTS target_amount;
