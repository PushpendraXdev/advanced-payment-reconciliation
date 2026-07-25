ALTER TABLE internal_transaction ALTER COLUMN status_of_payment SET DEFAULT 'pending';
UPDATE internal_transaction SET status_of_payment = 'pending' WHERE status_of_payment IS NULL;