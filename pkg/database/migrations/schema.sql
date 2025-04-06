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

CREATE TABLE IF NOT EXISTS "insurance_policies" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "policy_type" varchar(50),
  "start_date" timestamp,
  "end_date" timestamp,
  "premium_amount" numeric,
  "coverage_amount" numeric,
  "status" varchar(20)
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
CREATE INDEX IF NOT EXISTS idx_user_accounts_user_id ON "user_accounts" ("user_id");
CREATE INDEX IF NOT EXISTS idx_user_accounts_account_id ON "user_accounts" ("account_id");
CREATE INDEX IF NOT EXISTS idx_account_products_account_id ON "account_products" ("account_id");
CREATE INDEX IF NOT EXISTS idx_account_products_product_id ON "account_products" ("product_id");
CREATE INDEX IF NOT EXISTS idx_atm_transactions_atm_id ON "atm_transactions" ("atm_id");
CREATE INDEX IF NOT EXISTS idx_atm_transactions_account_id ON "atm_transactions" ("account_id");
CREATE INDEX IF NOT EXISTS idx_user_branches_user_id ON "user_branches" ("user_id");
CREATE INDEX IF NOT EXISTS idx_user_branches_branch_id ON "user_branches" ("branch_id");


DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_accounts_currency' AND table_name = 'accounts'
    ) THEN
        ALTER TABLE "accounts" 
        ADD CONSTRAINT FK_accounts_currency 
        FOREIGN KEY ("currency_id") REFERENCES "currency" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_refresh_tokens_user' AND table_name = 'refresh_tokens'
    ) THEN
        ALTER TABLE "refresh_tokens" 
        ADD CONSTRAINT FK_refresh_tokens_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_event_user' AND table_name = 'event'
    ) THEN
        ALTER TABLE "event" 
        ADD CONSTRAINT FK_event_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_users_role' AND table_name = 'users'
    ) THEN
        ALTER TABLE "users" 
        ADD CONSTRAINT FK_users_role
        FOREIGN KEY ("role_id") REFERENCES "roles" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_cards_account' AND table_name = 'cards'
    ) THEN
        ALTER TABLE "cards" 
        ADD CONSTRAINT FK_cards_account
        FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_transactions_from_account' AND table_name = 'transactions'
    ) THEN
        ALTER TABLE "transactions" 
        ADD CONSTRAINT FK_transactions_from_account
        FOREIGN KEY ("from_account") REFERENCES "accounts" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_transactions_to_account' AND table_name = 'transactions'
    ) THEN
        ALTER TABLE "transactions" 
        ADD CONSTRAINT FK_transactions_to_account
        FOREIGN KEY ("to_account") REFERENCES "accounts" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_user_settings_user' AND table_name = 'user_settings'
    ) THEN
        ALTER TABLE "user_settings" 
        ADD CONSTRAINT FK_user_settings_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_account_limits_account' AND table_name = 'account_limits'
    ) THEN
        ALTER TABLE "account_limits" 
        ADD CONSTRAINT FK_account_limits_account
        FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_staking_user' AND table_name = 'staking'
    ) THEN
        ALTER TABLE "staking" 
        ADD CONSTRAINT FK_staking_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_staking_currency' AND table_name = 'staking'
    ) THEN
        ALTER TABLE "staking" 
        ADD CONSTRAINT FK_staking_currency
        FOREIGN KEY ("currency_id") REFERENCES "currency" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_loans_user' AND table_name = 'loans'
    ) THEN
        ALTER TABLE "loans" 
        ADD CONSTRAINT FK_loans_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_loans_currency' AND table_name = 'loans'
    ) THEN
        ALTER TABLE "loans" 
        ADD CONSTRAINT FK_loans_currency
        FOREIGN KEY ("currency_id") REFERENCES "currency" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_bank_balance_currency' AND table_name = 'bank_balance'
    ) THEN
        ALTER TABLE "bank_balance" 
        ADD CONSTRAINT FK_bank_balance_currency
        FOREIGN KEY ("currency_id") REFERENCES "currency" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_cashback_card' AND table_name = 'cashback'
    ) THEN
        ALTER TABLE "cashback" 
        ADD CONSTRAINT FK_cashback_card
        FOREIGN KEY ("card_id") REFERENCES "cards" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_cashback_transaction' AND table_name = 'cashback'
    ) THEN
        ALTER TABLE "cashback" 
        ADD CONSTRAINT FK_cashback_transaction
        FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_insurance_policies_user' AND table_name = 'insurance_policies'
    ) THEN
        ALTER TABLE "insurance_policies" 
        ADD CONSTRAINT FK_insurance_policies_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_user_accounts_user' AND table_name = 'user_accounts'
    ) THEN
        ALTER TABLE "user_accounts" 
        ADD CONSTRAINT FK_user_accounts_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_user_accounts_account' AND table_name = 'user_accounts'
    ) THEN
        ALTER TABLE "user_accounts" 
        ADD CONSTRAINT FK_user_accounts_account
        FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_user_primary_address_user' AND table_name = 'user_primary_address'
    ) THEN
        ALTER TABLE "user_primary_address" 
        ADD CONSTRAINT FK_user_primary_address_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_account_products_account' AND table_name = 'account_products'
    ) THEN
        ALTER TABLE "account_products" 
        ADD CONSTRAINT FK_account_products_account
        FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_account_products_product' AND table_name = 'account_products'
    ) THEN
        ALTER TABLE "account_products" 
        ADD CONSTRAINT FK_account_products_product
        FOREIGN KEY ("product_id") REFERENCES "products" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_atm_transactions_atm' AND table_name = 'atm_transactions'
    ) THEN
        ALTER TABLE "atm_transactions" 
        ADD CONSTRAINT FK_atm_transactions_atm
        FOREIGN KEY ("atm_id") REFERENCES "atms" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_atm_transactions_account' AND table_name = 'atm_transactions'
    ) THEN
        ALTER TABLE "atm_transactions" 
        ADD CONSTRAINT FK_atm_transactions_account
        FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_branches_manager' AND table_name = 'branches'
    ) THEN
        ALTER TABLE "branches" 
        ADD CONSTRAINT FK_branches_manager
        FOREIGN KEY ("manager_id") REFERENCES "users" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_user_branches_user' AND table_name = 'user_branches'
    ) THEN
        ALTER TABLE "user_branches" 
        ADD CONSTRAINT FK_user_branches_user
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_user_branches_branch' AND table_name = 'user_branches'
    ) THEN
        ALTER TABLE "user_branches" 
        ADD CONSTRAINT FK_user_branches_branch
        FOREIGN KEY ("branch_id") REFERENCES "branches" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;
END
$$;


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