CREATE TABLE IF NOT EXISTS users
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login         VARCHAR(32) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);