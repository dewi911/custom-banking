DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_status') THEN
        CREATE TYPE "transaction_status" AS ENUM (
          'pending',
          'completed',
          'failed',
          'cancelled'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_type') THEN
        CREATE TYPE "event_type" AS ENUM (
          'login',
          'logout',
          'transfer',
          'password_change',
          'staking_start',
          'staking_end',
          'loan_request',
          'loan_approval',
          'loan_repayment'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'card_type') THEN
        CREATE TYPE "card_type" AS ENUM (
          'standard',
          'child',
          'senior',
          'premium'
        );
    END IF;
END
$$;



CREATE TABLE IF NOT EXISTS "roles" (
  "id" serial PRIMARY KEY,
  "name" varchar(30)
);

CREATE TABLE IF NOT EXISTS "users" (
  "id" serial PRIMARY KEY,
  "name" varchar(100),
  "surname" varchar(100),
  "email" varchar(100),
  "username" varchar(100),
  "password" varchar(255),
  "role_id" integer,
  "blocked" boolean DEFAULT false,
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "currency" (
  "id" serial PRIMARY KEY,
  "name" varchar(50),
  "code" varchar(3)
);

CREATE TABLE IF NOT EXISTS "accounts" (
  "id" serial PRIMARY KEY,
  "iban" varchar(100),
  "currency_id" integer,
  "blocked" boolean DEFAULT false,
  "amount" numeric DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "refresh_tokens" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "token" varchar(255),
  "expires_at" timestamp
);

CREATE TABLE IF NOT EXISTS "event" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "type" event_type,
  "metadata" jsonb,
  "time" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "cards" (
  "id" serial PRIMARY KEY,
  "account_id" integer,
  "card_number" varchar(16),
  "cardholder_name" varchar(250),
  "expiration_date" timestamp,
  "cvv_code" varchar(3),
  "card_type" card_type,
  "cashback_percentage" decimal(5,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "transactions" (
  "id" serial PRIMARY KEY,
  "from_account" integer,
  "to_account" integer,
  "amount" numeric,
  "status" transaction_status,
  "date_created" timestamp DEFAULT CURRENT_TIMESTAMP,
  "date_updated" timestamp DEFAULT CURRENT_TIMESTAMP,
  "transaction_type" varchar(50),
  "description" text,
  "reference_number" varchar(50)
);

CREATE TABLE IF NOT EXISTS "cashback" (
  "id" serial PRIMARY KEY,
  "card_id" integer,
  "transaction_id" integer,
  "amount" numeric,
  "date_created" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "audit_log" (
  "id" serial PRIMARY KEY,
  "table_name" varchar(50),
  "record_id" integer,
  "action" varchar(10),
  "changed_data" jsonb,
  "user_id" integer,
  "timestamp" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "user_settings" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "notification_preferences" jsonb,
  "language" varchar(10) DEFAULT 'en',
  "theme" varchar(20) DEFAULT 'light'
);

CREATE TABLE IF NOT EXISTS "account_limits" (
  "id" serial PRIMARY KEY,
  "account_id" integer,
  "daily_transfer_limit" numeric,
  "monthly_transfer_limit" numeric
);

CREATE TABLE IF NOT EXISTS "staking" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "amount" numeric,
  "currency_id" integer,
  "start_date" timestamp,
  "end_date" timestamp,
  "interest_rate" decimal(5,2),
  "status" varchar(20)
);

CREATE TABLE IF NOT EXISTS "staking_interests" (
    "id" serial PRIMARY KEY,
    "staking_id" integer REFERENCES staking(id),
    "amount" numeric,
    "date_calculated" timestamp DEFAULT CURRENT_TIMESTAMP,
    "description" text
);

CREATE TABLE IF NOT EXISTS "loans" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "amount" numeric,
  "currency_id" integer,
  "start_date" timestamp,
  "end_date" timestamp,
  "interest_rate" decimal(5,2),
  "status" varchar(20),
  "remaining_amount" numeric
);

CREATE TABLE IF NOT EXISTS "loan_payments" (
  "id" serial PRIMARY KEY,
  "loan_id" integer REFERENCES loans(id),
  "amount" numeric,
  "date" timestamp,
  "status" varchar(20),
  "payment_method" varchar(50),
  "transaction_id" integer
);

CREATE TABLE IF NOT EXISTS "insurance_policies" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "type" varchar(50),
  "policy_type" varchar(50),
  "insured_item" varchar(255),
  "coverage_amount" numeric,
  "premium_amount" numeric,
  "premium" numeric,
  "start_date" timestamp,
  "end_date" timestamp,
  "status" varchar(20),
  "policy_number" varchar(50),
  "description" text,
  "currency_id" integer,
  "payment_account_id" integer
);

CREATE TABLE IF NOT EXISTS "insurance_claims" (
  "id" serial PRIMARY KEY,
  "insurance_id" integer REFERENCES insurance_policies(id),
  "claim_date" timestamp,
  "description" text,
  "status" varchar(20),
  "amount" numeric,
  "filing_date" timestamp DEFAULT CURRENT_TIMESTAMP,
  "resolution_date" timestamp,
  "document_links" text
);

CREATE TABLE IF NOT EXISTS "bank_balance" (
  "id" serial PRIMARY KEY,
  "currency_id" integer,
  "total_amount" numeric,
  "available_amount" numeric,
  "staked_amount" numeric,
  "loaned_amount" numeric
);

CREATE TABLE IF NOT EXISTS "user_accounts" (
  "user_id" integer,
  "account_id" integer,
  "is_primary" boolean DEFAULT false,
  "access_level" varchar(20) DEFAULT 'owner',
  PRIMARY KEY ("user_id", "account_id")
);

CREATE TABLE IF NOT EXISTS "user_primary_address" (
  "user_id" integer PRIMARY KEY,
  "address_line1" varchar(100),
  "address_line2" varchar(100),
  "city" varchar(50),
  "state" varchar(50),
  "country" varchar(50),
  "postal_code" varchar(20)
);

CREATE TABLE IF NOT EXISTS "products" (
  "id" serial PRIMARY KEY,
  "name" varchar(100),
  "description" text,
  "type" varchar(50),
  "interest_rate" decimal(5,2),
  "min_balance" numeric,
  "max_balance" numeric
);

CREATE TABLE IF NOT EXISTS "account_products" (
  "account_id" integer,
  "product_id" integer,
  "start_date" timestamp,
  "end_date" timestamp,
  PRIMARY KEY ("account_id", "product_id")
);

CREATE TABLE IF NOT EXISTS "atms" (
  "id" serial PRIMARY KEY,
  "location_name" varchar(100),
  "address" varchar(255),
  "latitude" decimal(9,6),
  "longitude" decimal(9,6),
  "status" varchar(20)
);

CREATE TABLE IF NOT EXISTS "atm_transactions" (
  "id" serial PRIMARY KEY,
  "atm_id" integer,
  "account_id" integer,
  "transaction_type" varchar(20),
  "amount" numeric,
  "timestamp" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "branches" (
  "id" serial PRIMARY KEY,
  "name" varchar(100),
  "address" varchar(255),
  "phone" varchar(20),
  "manager_id" integer
);

CREATE TABLE IF NOT EXISTS "user_branches" (
  "user_id" integer,
  "branch_id" integer,
  "role" varchar(50),
  "start_date" date,
  "end_date" date,
  PRIMARY KEY ("user_id", "branch_id")
);


CREATE INDEX IF NOT EXISTS idx_users_email_username ON "users" ("email", "username");
CREATE INDEX IF NOT EXISTS idx_users_role_id ON "users" ("role_id");
CREATE INDEX IF NOT EXISTS idx_accounts_currency_id ON "accounts" ("currency_id");
CREATE INDEX IF NOT EXISTS idx_transactions_accounts ON "transactions" ("from_account", "to_account");
CREATE INDEX IF NOT EXISTS idx_transactions_date ON "transactions" ("date_created");
CREATE INDEX IF NOT EXISTS idx_cashback_card_id ON "cashback" ("card_id");
CREATE INDEX IF NOT EXISTS idx_cashback_transaction_id ON "cashback" ("transaction_id");
CREATE INDEX IF NOT EXISTS idx_loans_user_id ON "loans" ("user_id");
CREATE INDEX IF NOT EXISTS idx_loans_currency_id ON "loans" ("currency_id");
CREATE INDEX IF NOT EXISTS idx_insurance_user_id ON "insurance_policies" ("user_id");
CREATE INDEX IF NOT EXISTS idx_insurance_claims_insurance_id ON "insurance_claims" ("insurance_id");
CREATE INDEX IF NOT EXISTS idx_user_accounts_user_id ON "user_accounts" ("user_id");
CREATE INDEX IF NOT EXISTS idx_user_accounts_account_id ON "user_accounts" ("account_id");
CREATE INDEX IF NOT EXISTS idx_account_products_account_id ON "account_products" ("account_id");
CREATE INDEX IF NOT EXISTS idx_account_products_product_id ON "account_products" ("product_id");
CREATE INDEX IF NOT EXISTS idx_atm_transactions_atm_id ON "atm_transactions" ("atm_id");
CREATE INDEX IF NOT EXISTS idx_atm_transactions_account_id ON "atm_transactions" ("account_id");
CREATE INDEX IF NOT EXISTS idx_user_branches_user_id ON "user_branches" ("user_id");
CREATE INDEX IF NOT EXISTS idx_user_branches_branch_id ON "user_branches" ("branch_id");
CREATE INDEX IF NOT EXISTS idx_loan_payments_loan_id ON loan_payments (loan_id);
CREATE INDEX IF NOT EXISTS idx_staking_interests_staking_id ON "staking_interests" ("staking_id");


INSERT INTO roles (id, name)
VALUES (1, 'admin'), (2, 'user'), (3, 'anonymous')
ON CONFLICT (id) DO NOTHING;


INSERT INTO currency (id, name, code)
VALUES 
  (1, 'US Dollar', 'USD'),
  (2, 'Euro', 'EUR'),
  (3, 'British Pound', 'GBP'),
  (4, 'Japanese Yen', 'JPY'),
  (5, 'Russian Ruble', 'RUB')
ON CONFLICT (id) DO NOTHING;


CREATE OR REPLACE FUNCTION calculate_staking_interest()
RETURNS TRIGGER AS $$
DECLARE
interest_amount NUMERIC;
    last_calculation_date TIMESTAMP;
    days_since_last NUMERIC;
    daily_rate NUMERIC;
BEGIN
SELECT MAX(date_calculated) INTO last_calculation_date
FROM staking_interests
WHERE staking_id = NEW.id;

IF last_calculation_date IS NULL THEN
        last_calculation_date := NEW.start_date;
END IF;

    IF (CURRENT_TIMESTAMP - last_calculation_date < INTERVAL '1 day') OR
       NEW.status != 'active' THEN
        RETURN NEW;
END IF;

    days_since_last := EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - last_calculation_date)) / 86400;

    daily_rate := NEW.interest_rate / 36500;

    interest_amount := NEW.amount * daily_rate * days_since_last;

INSERT INTO staking_interests (
    staking_id,
    amount,
    date_calculated,
    description
) VALUES (
             NEW.id,
             interest_amount,
             CURRENT_TIMESTAMP,
             'Automatic daily interest calculation'
         );

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trigger_calculate_staking_interest
AFTER UPDATE ON staking
                    FOR EACH ROW
                    EXECUTE FUNCTION calculate_staking_interest();

CREATE OR REPLACE FUNCTION daily_interest_calculation()
RETURNS void AS $$
DECLARE
staking_rec RECORD;
BEGIN
FOR staking_rec IN SELECT * FROM staking WHERE status = 'active' LOOP
-- init trigger
UPDATE staking
SET id = staking_rec.id
WHERE id = staking_rec.id;
END LOOP;
END;
$$ LANGUAGE plpgsql;

