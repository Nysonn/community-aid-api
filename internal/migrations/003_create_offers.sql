-- +goose Up
CREATE TABLE offers (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id        UUID          NOT NULL REFERENCES emergency_requests(id) ON DELETE CASCADE,
    responder_name    TEXT          NOT NULL,
    responder_contact TEXT          NOT NULL,
    offer_type        TEXT          NOT NULL
                                    CHECK (offer_type IN ('transport', 'donation', 'expertise')),
    status            TEXT          NOT NULL DEFAULT 'pending'
                                    CHECK (status IN ('pending', 'accepted', 'fulfilled')),
    latitude          NUMERIC(10,7),
    longitude         NUMERIC(10,7),
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE offers;
