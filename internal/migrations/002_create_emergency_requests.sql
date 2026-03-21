-- +goose Up
CREATE TABLE emergency_requests (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         TEXT          NOT NULL,
    description   TEXT          NOT NULL,
    type          TEXT          NOT NULL
                                CHECK (type IN ('medical', 'food', 'rescue', 'shelter')),
    status        TEXT          NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending', 'approved', 'rejected', 'closed')),
    location_name TEXT          NOT NULL,
    latitude      NUMERIC(10,7),
    longitude     NUMERIC(10,7),
    media_urls    TEXT[]        NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE emergency_requests;
