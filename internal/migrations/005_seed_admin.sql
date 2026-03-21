-- +goose Up
INSERT INTO users (clerk_id, full_name, email, phone_number, role, is_active)
VALUES ('admin_seed', 'CommunityAid Admin', 'admin@communityaid.com', '0000000000', 'admin', TRUE)
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE email = 'admin@communityaid.com';
