-- +goose Up
ALTER TABLE offers
    ADD COLUMN IF NOT EXISTS expertise_details TEXT,
    ADD COLUMN IF NOT EXISTS vehicle_type TEXT,
    ADD COLUMN IF NOT EXISTS donation_amount NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS payment_method TEXT CHECK (payment_method IN ('mobile_money', 'visa')),
    ADD COLUMN IF NOT EXISTS mobile_money_provider TEXT CHECK (mobile_money_provider IN ('airtel_money', 'mtn_momo')),
    ADD COLUMN IF NOT EXISTS mobile_money_number_masked TEXT,
    ADD COLUMN IF NOT EXISTS card_last4 TEXT,
    ADD COLUMN IF NOT EXISTS card_expiry_month INTEGER,
    ADD COLUMN IF NOT EXISTS card_expiry_year INTEGER,
    ADD COLUMN IF NOT EXISTS cardholder_name TEXT;

-- +goose Down
ALTER TABLE offers
    DROP COLUMN IF EXISTS cardholder_name,
    DROP COLUMN IF EXISTS card_expiry_year,
    DROP COLUMN IF EXISTS card_expiry_month,
    DROP COLUMN IF EXISTS card_last4,
    DROP COLUMN IF EXISTS mobile_money_number_masked,
    DROP COLUMN IF EXISTS mobile_money_provider,
    DROP COLUMN IF EXISTS payment_method,
    DROP COLUMN IF EXISTS donation_amount,
    DROP COLUMN IF EXISTS vehicle_type,
    DROP COLUMN IF EXISTS expertise_details;