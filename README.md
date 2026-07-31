# 💳 Advanced Payment Reconciliation System

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white" />
  <img src="https://img.shields.io/badge/Gin-Framework-00ADD8?style=for-the-badge" />
  <img src="https://img.shields.io/badge/Docker-Containerized-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
  <img src="https://img.shields.io/badge/Tests-Passing-brightgreen?style=for-the-badge" />
</p>

<p align="center">
  <b>An automated, human-in-the-loop payment reconciliation backend</b><br/>
  Ingests transactions from two independent sources, matches them automatically on a schedule,<br/>
  and routes every match through an approval gate before it's considered final.
</p>

---

## 📖 Why This Exists

Payment reconciliation is the process every fintech company runs to verify that money recorded **internally** actually matches what a **payment gateway or bank** processed — catching duplicates, missing payments, and mismatches before they become accounting errors.

This project simulates that entire pipeline — ingestion, automated matching, human approval, audit trail, and reporting — the same way production fintech systems are architected, without relying on a third-party reconciliation SaaS.

---

## 🏗️ Architecture

```mermaid
flowchart TB
    subgraph Sources["📥 Data Sources"]
        A[Internal Order System]
        B[Payment Gateway Webhook]
    end

    subgraph Ingestion["Ingestion Layer"]
        C[POST /transactions/internal]
        D[POST /gateway/webhook<br/><i>idempotency-protected</i>]
    end

    subgraph Engine["⚙️ Automated Matching Engine"]
        E[Scheduled Background Worker<br/><i>goroutine + ticker, runs every 30s</i>]
        F[Hash-Map Matcher<br/><i>O(n+m) amount-based matching</i>]
    end

    subgraph Approval["🧑‍💼 Human Approval Gate"]
        G{pending_approval queue}
        H[Approve<br/>single or batch]
        I[Reject<br/>unlinks both sides]
    end

    subgraph Records["📜 Records"]
        J[(matches)]
        K[(audit_logs)]
    end

    L[✅ matched]

    A --> C --> J
    B --> D --> J
    E --> F --> J
    J --> G
    G --> H --> L
    G --> I --> A
    H --> K
    I --> K

    style Sources fill:#1a1a2e,stroke:#0f3460,color:#fff
    style Ingestion fill:#16213e,stroke:#0f3460,color:#fff
    style Engine fill:#0f3460,stroke:#e94560,color:#fff
    style Approval fill:#533483,stroke:#e94560,color:#fff
    style Records fill:#1a1a2e,stroke:#0f3460,color:#fff
    style L fill:#0f9d58,stroke:#0a7c3f,color:#fff
```

**The key idea:** ingestion, matching, flagging, and logging are **100% automated** — no manual trigger required in normal operation. The only human action in the entire pipeline is a lightweight approve/reject decision on already-matched transactions, right before they're considered final. This mirrors how real reconciliation platforms (Razorpay, HighRadius, etc.) balance automation with financial risk control.

---

## 🔄 Transaction Lifecycle

```mermaid
sequenceDiagram
    participant Internal as Internal System
    participant Gateway as Payment Gateway
    participant DB as PostgreSQL
    participant Engine as Matching Engine
    participant Human as Approver

    Internal->>DB: Create transaction (status: pending)
    Gateway->>DB: Webhook event (idempotency key checked)
    Note over Engine: Runs automatically every 30s
    Engine->>DB: Fetch pending internal + unlinked gateway txns
    Engine->>Engine: Hash-map match on amount
    Engine->>DB: Create match (status: pending_approval)
    Engine->>DB: Write audit log entry
    Human->>DB: Approve (single or batch)
    DB->>DB: status → matched
    DB->>DB: Write audit log entry
```

---

## 🧰 Tech Stack

| Layer | Technology | Why |
|---|---|---|
| **Language** | Go 1.25 | Goroutines + channels for concurrent, scheduled matching without external job queues |
| **Database** | PostgreSQL 16 (`pgx` driver) | ACID transactions, `DECIMAL` for exact currency precision |
| **Web Framework** | Gin | Lightweight routing, JSON binding, middleware |
| **Migrations** | golang-migrate / raw SQL | Versioned, reproducible schema changes |
| **Containerization** | Docker + Docker Compose | Multi-stage build (`CGO_ENABLED=0`) for a minimal, portable image |
| **Testing** | Go `testing` package | Table-driven tests covering normal, edge, and adversarial cases |

---

## 🗄️ Database Schema

```mermaid
erDiagram
    internal_transaction ||--o{ matches : "matched via"
    gateway_transactions ||--o{ matches : "matched via"
    matches ||--o{ audit_logs : "logged by"
    internal_transaction ||--o{ audit_logs : "logged by"
    gateway_transactions ||--o{ audit_logs : "logged by"

    internal_transaction {
        int id PK
        int order_id
        decimal amount
        string mode_of_payment
        string status_of_payment
        timestamp created_at
    }
    gateway_transactions {
        int id PK
        int internal_transaction_id FK
        string payment_id UK
        int user_id
        decimal amount
        string utr
        string mode_of_payment
        string gateway_name
        string status
        timestamp created_at
    }
    matches {
        int id PK
        int internal_transaction_id FK
        int gateway_transaction_id FK
        string status
        timestamp matched_at
        timestamp approved_at
        string approved_by
    }
    audit_logs {
        int id PK
        int matched_id FK
        int internal_transaction_id FK
        int gateway_transaction_id FK
        string action
        string status
        text details
        timestamp created_at
    }
```

---

## 🔌 API Reference

### Ingestion
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/transactions/internal` | Create an internal transaction |
| `GET` | `/transactions` | List all internal transactions |
| `POST` | `/gateway/webhook` | Receive a payment-gateway event (idempotency-protected) |

### Reconciliation
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/reconcile/run` | Manually trigger a reconciliation cycle *(also runs automatically every 30s)* |
| `GET` | `/matches/pending-approval` | List matches awaiting human approval |
| `POST` | `/matches/:id/approve` | Approve a single match → posts to ledger |
| `POST` | `/matches/approve-batch` | Approve **all** pending matches in one call |
| `POST` | `/matches/:id/reject` | Reject a match and unlink both sides |

### Reporting
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/reports/summary` | Status breakdown + match-rate percentage |

<details>
<summary><b>Example: creating a transaction pair and reconciling</b></summary>

```bash
# 1. Create an internal transaction
curl -X POST http://localhost:8080/transactions/internal \
  -H "Content-Type: application/json" \
  -d '{"order_id": 501, "amount": 999.00, "mode_of_payment": "UPI"}'

# 2. Simulate the gateway confirming the same payment
curl -X POST http://localhost:8080/gateway/webhook \
  -H "Content-Type: application/json" \
  -d '{"payment_id": "PAY501", "user_id": 1, "amount": 999.00, "utr": "UTR501", "mode_of_payment": "UPI", "gateway_name": "Razorpay"}'

# 3. Trigger reconciliation (or just wait — it runs automatically)
curl -X POST http://localhost:8080/reconcile/run

# 4. Approve the match
curl -X POST http://localhost:8080/matches/1/approve
```
</details>

---

## 🧠 Matching Algorithm

Matching is done **in-memory** using a hash map keyed on amount, converted to integer paise (`int64(amount * 100)`) to sidestep floating-point precision issues:

```go
gatewayMap := make(map[int64][]GatewayTxn)   // handles duplicate amounts safely
for _, g := range gatewayTxns {
    key := int64(g.Amount * 100)
    gatewayMap[key] = append(gatewayMap[key], g)
}
```

This gives **O(n + m)** matching instead of an **O(n × m)** nested-loop comparison — the difference between ~2,000 operations and ~1,000,000 for two lists of 1,000 transactions each.

---

## 🛡️ Why an Approval Gate?

```mermaid
flowchart LR
    A[Match Found] -->|Automated| B[pending_approval]
    B -->|Human clicks Approve| C[✅ matched / ledger-ready]
    B -->|Human clicks Reject| D[🔄 unlinked, back to pending]

    style A fill:#0f3460,color:#fff
    style B fill:#533483,color:#fff
    style C fill:#0f9d58,color:#fff
    style D fill:#e94560,color:#fff
```

Real reconciliation platforms don't post matched transactions straight to the books, even when matching itself is fully automated — they insert one human checkpoint right before the ledger write, since money is being committed to permanent records. This project follows the same pattern: **>90% of the pipeline is automated**, and the only manual step is a lightweight approve/reject click on already-matched transactions.

---

## ✅ Testing

Table-driven tests cover the core matching engine — normal matches, no-match cases, duplicate amounts, empty inputs, partial matches, zero/negative amounts, and multi-transaction scenarios:

```bash
go test -v ./internal/matcher/...
```

---

## 🐳 Running with Docker

```bash
docker-compose up --build
```

This spins up two containers — the Go app and PostgreSQL — on an isolated network, with the app connecting to the database via its service name (`db`), not `localhost`. The Go binary is built with a multi-stage Dockerfile (`CGO_ENABLED=0`) so the final image needs nothing but the compiled binary.

Once the containers are up, apply the schema:

```bash
psql -h localhost -U postgres -d reconciliation_db -f internal/db/migrations/<file>.up.sql
# repeat for each migration file, in order
```

---

## 🚀 Local Setup (without Docker)

1. Clone the repo and install dependencies:
   ```bash
   go mod tidy
   ```
2. Create a `.env` file:
   ```
   DATABASE_URL=postgres://postgres:yourpassword@localhost:5432/reconciliation_db?sslmode=disable
   ```
3. Run the migrations in `internal/db/migrations/` against your local Postgres.
4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

---

## 📈 Project Status

| Phase | Description | Status |
|---|---|---|
| 1 | Foundations (Go + Gin + Postgres) | ✅ |
| 2 | Ingestion (internal + gateway, idempotency) | ✅ |
| 3 | Matching engine (hash-map, O(n+m)) | ✅ |
| 4 | Automation (background scheduler) | ✅ |
| 5 | Approval gate (single, batch, reject) | ✅ |
| 6 | Audit logging | ✅ |
| 7 | Reporting (status breakdown, match rate) | ✅ |
| 8 | Testing + Docker | ✅ |

**Possible future extensions:** event-driven ingestion via RabbitMQ/Kafka, a minimal frontend dashboard, tolerance-based fuzzy matching for partial payments.

---

<p align="center"><i>Built as a deep, production-adjacent portfolio project — not a tutorial clone.</i></p>