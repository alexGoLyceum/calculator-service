CREATE TABLE IF NOT EXISTS tasks
(
    id             UUID PRIMARY KEY NOT NULL,
    expression_id  UUID             NOT NULL,
    arg1_value     DOUBLE PRECISION,
    arg1_task_id   UUID,
    arg2_value     DOUBLE PRECISION,
    arg2_task_id   UUID,
    operator       VARCHAR(10)      NOT NULL,
    operation_time TIMESTAMP        NOT NULL,
    final_task     BOOLEAN          NOT NULL DEFAULT FALSE,
    status         VARCHAR(50)      NOT NULL DEFAULT 'pending',
    result         DOUBLE PRECISION          DEFAULT NULL,
    created_at     TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (expression_id) REFERENCES expressions (id) ON DELETE CASCADE,
    FOREIGN KEY (arg1_task_id) REFERENCES tasks (id) ON DELETE SET NULL,
    FOREIGN KEY (arg2_task_id) REFERENCES tasks (id) ON DELETE SET NULL
);