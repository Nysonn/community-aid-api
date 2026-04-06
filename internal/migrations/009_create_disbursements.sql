-- +goose Up
CREATE TABLE disbursements (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    offer_id                 UUID         NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
    request_id               UUID         NOT NULL REFERENCES emergency_requests(id) ON DELETE CASCADE,
    donor_name               TEXT         NOT NULL,
    donor_email              TEXT         NOT NULL,
    amount                   NUMERIC(12,2) NOT NULL,
    recipient_name           TEXT         NOT NULL,
    -- payment details copied from the request at time of creation
    payment_type             TEXT         NOT NULL CHECK (payment_type IN ('bank', 'mobile_money')),
    bank_account_name        TEXT,
    bank_account_number      TEXT,
    bank_name                TEXT,
    receiving_mobile_provider TEXT,
    receiving_mobile_number  TEXT,
    status                   TEXT         NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'disbursed')),
    disbursed_at             TIMESTAMPTZ,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE disbursements;
