CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    internal_transaction_id INT REFERENCES internal_transaction(id),
    gateway_transaction_id INT REFERENCES gateway_transactions(id),
    status VARCHAR(50) DEFAULT 'pending_approval',
    matched_at TIMESTAMP DEFAULT NOW(),
    approved_at TIMESTAMP,
    approved_by VARCHAR(100)
);