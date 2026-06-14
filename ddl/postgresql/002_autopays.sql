-- 002_autopays.sql — fixed recurring monthly commitments (EMIs, subscriptions,
-- insurance, SIPs). Only rows with status 'active' count toward committed money.
-- confidence_score is populated only for entries detected from email
-- (source='email_auto') that are awaiting user confirmation.

CREATE TABLE IF NOT EXISTS autopays (
    id                BIGSERIAL     PRIMARY KEY,
    user_id           BIGINT        NOT NULL,
    name              VARCHAR(120)  NOT NULL,
    type              VARCHAR(32)   NOT NULL,
    amount            NUMERIC(12,2) NOT NULL,
    deduct_day        INT           NOT NULL CHECK (deduct_day BETWEEN 1 AND 31),
    source            VARCHAR(16)   NOT NULL DEFAULT 'manual',
    status            VARCHAR(16)   NOT NULL DEFAULT 'active',
    confidence_score  DOUBLE PRECISION,
    notes             TEXT          NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_autopays_user_status ON autopays (user_id, status);
