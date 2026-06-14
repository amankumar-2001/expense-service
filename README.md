# expense-service (a.k.a. autopay-service)

The KharchiBook expense + autopay backend. It owns expenses, recurring
commitments (EMIs / subscriptions / insurance / SIPs), the salary inputs, and the
derived "free money" analytics the web dashboard renders.

It is a sibling of `auth-service` and deliberately mirrors its architecture:
Gin + GORM (PostgreSQL) + go-redis, hand-wired DI (`pkg/di`), per-env JSON config
via Viper, the `platlogger` slog wrapper, and the typed HTTP error envelope.

## Identity

This service does **not** issue tokens. It only **verifies** the RS256 access
tokens minted by `auth-service`, using that service's **public** key
(`assets/keys/jwt_public.pem`). The verified `user_id` (the JWT `sub`, an int64)
scopes every query. There is no cross-database foreign key to the auth `users`
table.

## Data & infra

- **Postgres** — same setup as auth-service, a **separate database**
  (`expense_service`). Tables: `expenses`, `autopays`, `user_finance`
  (`ddl/postgresql/*.sql`; GORM `AutoMigrate` runs them in dev).
- **Redis** — the **same instance** auth-service uses, on a distinct DB index
  (`1`) to isolate the keyspace. Caches the committed-money summary with a short
  TTL, invalidated on any autopay/salary change.

## API (all under `/v1`, JWT required)

| Method | Path | Purpose |
|--------|------|---------|
| POST   | `/v1/expenses` | Log an expense |
| GET    | `/v1/expenses` | List expenses (`?month=YYYY-MM&category=`) |
| DELETE | `/v1/expenses/last` | Delete the last logged expense |
| GET    | `/v1/expenses/summary` | Monthly spend by category (`?month=`) |
| GET    | `/v1/autopays` | List commitments (`?status=&type=`) |
| POST   | `/v1/autopays` | Add a manual commitment |
| PATCH  | `/v1/autopays/:id` | Update a commitment |
| DELETE | `/v1/autopays/:id` | Soft-delete (cancel) a commitment |
| POST   | `/v1/autopays/:id/confirm` | Confirm an auto-detected commitment |
| PUT    | `/v1/salary` | Set monthly salary + salary day |
| GET    | `/v1/analytics/committed` | Committed / free-money breakdown |
| GET    | `/v1/analytics/upcoming` | Deductions due soon (`?days=7`) |

Plus `GET /healthz` (liveness) and `GET /readyz` (pings Postgres + Redis).

The JSON field names match the web client's `types.ts` exactly, so the dashboard
works once `NEXT_PUBLIC_MOCK_EXPENSE` is turned off and
`NEXT_PUBLIC_EXPENSE_URL` points here (default `http://localhost:8082`).

## Run locally

Local dev reuses **auth-service's Postgres + Redis** — one Postgres instance
hosts both `auth_service` and `expense_service`, and Redis is shared (expense
uses DB index 1). On this machine those run via Homebrew (`postgresql@16` on
:5432, `redis` on :6379), matching `assets/kharchibook/dev-config.json`
(role `auth`, password `auth`).

```bash
# 1. One-time: create the expense_service database (owned by the auth role).
createdb -h localhost -p 5432 -O auth expense_service
#    (or, if Postgres runs via docker: docker compose up -d)

# 2. The JWT public key — must be the public key of whichever auth-service
#    issues the tokens you log in with. For the DEPLOYED auth-service:
curl -s https://auth-service-p7kn.onrender.com/v1/public/auth/.well-known/public-key \
  -o assets/keys/jwt_public.pem
#    (for a LOCAL auth-service: cp ../auth-service/assets/keys/jwt_public.pem assets/keys/)

# 3. Run it (listens on :8082). AutoMigrate creates the tables on first boot.
make run        # or: VS Code → "Run expense-service (local · brew PG/Redis)"
```

## Out of scope (later phases)

Gmail OAuth + the AI auto-detection funnel, WhatsApp webhooks, Razorpay billing,
cron reminder workers, and natural-language expense parsing are deferred — see
`PRD/KHARCHIBOOK_PLATFORM_PRD.md`. This service currently implements the
expense + autopay + analytics core that powers the dashboard.
