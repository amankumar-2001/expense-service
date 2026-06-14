-- 003_user_finance.sql — per-user salary inputs that drive the free-money
-- calculation. One row per auth-service user (keyed by user_id).

CREATE TABLE IF NOT EXISTS user_finance (
    user_id        BIGINT        PRIMARY KEY,
    monthly_salary NUMERIC(12,2) NOT NULL DEFAULT 0,
    salary_day     INT           NOT NULL DEFAULT 1 CHECK (salary_day BETWEEN 1 AND 31),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);
