-- 001_expenses.sql — logged discretionary expenses.
-- user_id references the auth-service user (learned from the JWT). There is no
-- cross-database foreign key; ownership is enforced in application queries.

CREATE TABLE IF NOT EXISTS expenses (
    id            BIGSERIAL     PRIMARY KEY,
    user_id       BIGINT        NOT NULL,
    amount        NUMERIC(12,2) NOT NULL,
    category      VARCHAR(32)   NOT NULL,
    note          TEXT          NOT NULL DEFAULT '',
    raw_text      TEXT          NOT NULL DEFAULT '',
    expense_date  DATE          NOT NULL,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_expenses_user_date ON expenses (user_id, expense_date);
