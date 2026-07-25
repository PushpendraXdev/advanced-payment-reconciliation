CREATE TABLE gateway_transactions (
    id SERIAL PRIMARY KEY,
    internal_transaction_id INT REFERENCES internal_transaction(id),
    payment_id VARCHAR(100) UNIQUE NOT NULL,
    user_id INT,
    amount DECIMAL(12,2) NOT NULL,
    utr VARCHAR(50),
    mode_of_payment VARCHAR(50),
    gateway_name VARCHAR(50),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);