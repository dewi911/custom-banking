DROP TABLE IF EXISTS insurance_policies CASCADE;

CREATE TABLE IF NOT EXISTS insurance_policies (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    insured_item VARCHAR(255) NOT NULL,
    coverage_amount NUMERIC(15,2) NOT NULL,
    premium NUMERIC(15,2) NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL,
    policy_number VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    currency_id BIGINT NOT NULL REFERENCES currency(id),
    payment_account_id BIGINT NOT NULL REFERENCES accounts(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_insurance_policies_user_id ON insurance_policies(user_id);
CREATE INDEX IF NOT EXISTS idx_insurance_policies_status ON insurance_policies(status);
CREATE INDEX IF NOT EXISTS idx_insurance_policies_type ON insurance_policies(type);
CREATE INDEX IF NOT EXISTS idx_insurance_policies_policy_number ON insurance_policies(policy_number);

CREATE TABLE IF NOT EXISTS insurance_claims (
    id BIGSERIAL PRIMARY KEY,
    insurance_id BIGINT NOT NULL REFERENCES insurance_policies(id) ON DELETE CASCADE,
    claim_date TIMESTAMP NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(50) NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    filing_date TIMESTAMP NOT NULL,
    resolution_date TIMESTAMP,
    document_links TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_insurance_claims_insurance_id ON insurance_claims(insurance_id);
CREATE INDEX IF NOT EXISTS idx_insurance_claims_status ON insurance_claims(status);

CREATE OR REPLACE FUNCTION update_insurance_policies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_insurance_policies_updated_at
BEFORE UPDATE ON insurance_policies
FOR EACH ROW
EXECUTE FUNCTION update_insurance_policies_updated_at();

CREATE OR REPLACE FUNCTION update_insurance_claims_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_insurance_claims_updated_at
BEFORE UPDATE ON insurance_claims
FOR EACH ROW
EXECUTE FUNCTION update_insurance_claims_updated_at(); 