CREATE TABLE IF NOT EXISTS loan_payments (
    id serial PRIMARY KEY,
    loan_id integer REFERENCES loans(id),
    amount numeric,
    date timestamp,
    status varchar(20),
    payment_method varchar(50),
    transaction_id integer
    );

CREATE INDEX IF NOT EXISTS idx_loan_payments_loan_id ON loan_payments (loan_id);