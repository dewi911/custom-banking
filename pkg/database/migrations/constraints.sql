DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_accounts_currency' AND table_name = 'accounts'
    ) INTO constraint_exists;
    
    IF constraint_exists THEN
        RAISE NOTICE 'Dropping constraint FK_accounts_currency';
        EXECUTE 'ALTER TABLE "accounts" DROP CONSTRAINT "FK_accounts_currency"';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_refresh_tokens_user' AND table_name = 'refresh_tokens'
    ) INTO constraint_exists;
    
    IF constraint_exists THEN
        RAISE NOTICE 'Dropping constraint FK_refresh_tokens_user';
        EXECUTE 'ALTER TABLE "refresh_tokens" DROP CONSTRAINT "FK_refresh_tokens_user"';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_event_user' AND table_name = 'event'
    ) INTO constraint_exists;
    
    IF constraint_exists THEN
        RAISE NOTICE 'Dropping constraint FK_event_user';
        EXECUTE 'ALTER TABLE "event" DROP CONSTRAINT "FK_event_user"';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_users_role' AND table_name = 'users'
    ) INTO constraint_exists;
    
    IF constraint_exists THEN
        RAISE NOTICE 'Dropping constraint FK_users_role';
        EXECUTE 'ALTER TABLE "users" DROP CONSTRAINT "FK_users_role"';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_cards_account' AND table_name = 'cards'
    ) INTO constraint_exists;
    
    IF constraint_exists THEN
        RAISE NOTICE 'Dropping constraint FK_cards_account';
        EXECUTE 'ALTER TABLE "cards" DROP CONSTRAINT "FK_cards_account"';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_accounts_currency_1' AND table_name = 'accounts'
    ) INTO constraint_exists;
    
    IF NOT constraint_exists THEN
        RAISE NOTICE 'Adding constraint FK_accounts_currency_1';
        ALTER TABLE "accounts" 
        ADD CONSTRAINT FK_accounts_currency_1
        FOREIGN KEY ("currency_id") REFERENCES "currency" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    ELSE
        RAISE NOTICE 'Constraint FK_accounts_currency_1 already exists';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_refresh_tokens_user_1' AND table_name = 'refresh_tokens'
    ) INTO constraint_exists;
    
    IF NOT constraint_exists THEN
        RAISE NOTICE 'Adding constraint FK_refresh_tokens_user_1';
        ALTER TABLE "refresh_tokens" 
        ADD CONSTRAINT FK_refresh_tokens_user_1
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    ELSE
        RAISE NOTICE 'Constraint FK_refresh_tokens_user_1 already exists';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_event_user_1' AND table_name = 'event'
    ) INTO constraint_exists;
    
    IF NOT constraint_exists THEN
        RAISE NOTICE 'Adding constraint FK_event_user_1';
        ALTER TABLE "event" 
        ADD CONSTRAINT FK_event_user_1
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    ELSE
        RAISE NOTICE 'Constraint FK_event_user_1 already exists';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_users_role_1' AND table_name = 'users'
    ) INTO constraint_exists;
    
    IF NOT constraint_exists THEN
        RAISE NOTICE 'Adding constraint FK_users_role_1';
        ALTER TABLE "users" 
        ADD CONSTRAINT FK_users_role_1
        FOREIGN KEY ("role_id") REFERENCES "roles" ("id") 
        ON DELETE SET NULL ON UPDATE CASCADE;
    ELSE
        RAISE NOTICE 'Constraint FK_users_role_1 already exists';
    END IF;
END
$$;

DO $$
DECLARE
    constraint_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'FK_cards_account_1' AND table_name = 'cards'
    ) INTO constraint_exists;
    
    IF NOT constraint_exists THEN
        RAISE NOTICE 'Adding constraint FK_cards_account_1';
        ALTER TABLE "cards" 
        ADD CONSTRAINT FK_cards_account_1
        FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") 
        ON DELETE CASCADE ON UPDATE CASCADE;
    ELSE
        RAISE NOTICE 'Constraint FK_cards_account_1 already exists';
    END IF;
END
$$;