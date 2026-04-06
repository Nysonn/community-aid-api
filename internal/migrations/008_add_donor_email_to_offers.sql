-- +goose Up
ALTER TABLE offers
    ADD COLUMN donor_email TEXT;

-- +goose Down
ALTER TABLE offers
    DROP COLUMN IF EXISTS donor_email;
