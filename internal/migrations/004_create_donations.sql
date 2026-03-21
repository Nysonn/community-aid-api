-- +goose Up
CREATE TABLE donations (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID          NOT NULL REFERENCES emergency_requests(id) ON DELETE CASCADE,
    donor_name TEXT          NOT NULL,
    amount     NUMERIC(12,2) NOT NULL,
    date       DATE          NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE donations;
