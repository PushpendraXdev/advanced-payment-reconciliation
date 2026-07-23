CREATE TABLE internal_transaction (
   Id Serial PRIMARY KEY,
   Order_Id INT,
   Amount Decimal(12,2),
   Mode_OF_Payment VARCHAR(50),
   Status_Of_Payment VARCHAR(50),
  Created_At TIMESTAMP DEFAULT NOW()
);




-- psql -U postgres -d reconciliation_db -f internal/db/migrations/000001_create_internal_transactions_table.up.sql
-- psql -U postgres -d reconciliation_db
-- \dt
-- \d internal_transaction