-- +goose Up
CREATE TABLE users (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_id          TEXT        NOT NULL UNIQUE,
    full_name         TEXT        NOT NULL,
    email             TEXT        NOT NULL UNIQUE,
    phone_number      TEXT        NOT NULL,
    bio               TEXT,
    profile_image_url TEXT,
    role              TEXT        NOT NULL DEFAULT 'community_member'
                                  CHECK (role IN ('community_member', 'admin')),
    is_active         BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
