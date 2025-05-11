CREATE TABLE IF NOT EXISTS expressions
(
    id         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL,
    expression VARCHAR(255) NOT NULL,
    status     VARCHAR(50)  NOT NULL DEFAULT 'pending',
    result     DOUBLE PRECISION,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);