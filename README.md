# Advanced Payment Reconciliation System

An automated payment reconciliation backend built in Go + PostgreSQL. It ingests transactions from two independent sources (an internal order system and a simulated payment gateway), automatically matches them on a schedule, and routes matches through a human approval gate before they're considered final — modeled after real fintech reconciliation pipelines.

## Why this project

Payment reconciliation is the process of verifying that payments recorded internally match what a payment gateway/bank actually processed — catching duplicates, missing transactions, and mismatches before they become accounting errors. This project simulates that pipeline end-to-end: ingestion, automated matching, and an approval workflow, without relying on a third-party reconciliation SaaS.

## Architecture

```
Internal orders ──┐
                   ├─→ Matching Engine (scheduled, automatic) ─→ pending_approval
Gateway webhook ───┘                                                   │
                                                              Human approval
                                                                         │
                                                              matched / rejected
```

- **Ingestion**: two independent sources write to two separate tables, each with idempotency protection.
- **Matching**: a background scheduler (goroutine + ticker) runs automatically at a fixed interval — no manual trigger required in normal operation.
- **Approval gate**: matches are not finalized automatically. High-confidence matches sit in a `pending_approval` queue and are approved (individually or in bulk) or rejected by a human before being marked `matched`.

## Tech stack

- **Go** — goroutines + channels for concurrent, scheduled matching
- **PostgreSQL** (via `pgx`) — transactional storage, `DECIMAL` for exact currency precision
- **Gin** — HTTP routing
- **golang-migrate** — versioned schema migrations
- **godotenv** — environment-based configuration

## Database schema

| Table | Purpose |
|---|---|
| `internal_transaction` | Orders created by the internal system |
| `gateway_transactions` | Payment events received via webhook (idempotency-protected) |
| `matches` | Tracks each matched pair through `pending_approval` → `approved`/`rejected` |

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Health check |
| POST | `/transactions/internal` | Create an internal transaction |
| GET | `/transactions` | List all internal transactions |
| POST | `/gateway/webhook` | Receive a simulated gateway payment event |
| POST | `/reconcile/run` | Manually trigger a reconciliation cycle (also runs automatically on a schedule) |
| GET | `/matches/pending-approval` | List matches awaiting approval |
| POST | `/matches/:id/approve` | Approve a single match |
| POST | `/matches/approve-batch` | Approve all pending matches at once |
| POST | `/matches/:id/reject` | Reject a match and unlink both sides |

## Matching logic

Matching is done in-memory using a hash map keyed on amount (converted to integer paise to avoid floating-point precision issues), giving O(n + m) matching instead of an O(n × m) nested-loop comparison.

## Status

This project is under active development. Completed so far:
- [x] Ingestion (internal + gateway, with idempotency)
- [x] Exact-amount matching engine
- [x] Automated background scheduler
- [x] Approval gate (single, batch, reject)

Planned next:
- [ ] Audit logging
- [ ] Discrepancy handling for unmatched/partial transactions
- [ ] Dockerization
- [ ] Automated tests

## Setup

1. Clone the repo, `go mod tidy`
2. Set up a `.env` file with `DATABASE_URL`
3. Run migrations in `internal/db/migrations/`
4. `go run cmd/server/main.go`

---
*This README will be expanded as the project progresses.*
