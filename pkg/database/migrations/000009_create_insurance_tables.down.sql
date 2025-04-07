DROP TRIGGER IF EXISTS trigger_update_insurance_claims_updated_at ON insurance_claims;
DROP FUNCTION IF EXISTS update_insurance_claims_updated_at();

DROP TRIGGER IF EXISTS trigger_update_insurance_policies_updated_at ON insurance_policies;
DROP FUNCTION IF EXISTS update_insurance_policies_updated_at();

DROP INDEX IF EXISTS idx_insurance_claims_status;
DROP INDEX IF EXISTS idx_insurance_claims_insurance_id;

DROP INDEX IF EXISTS idx_insurance_policies_policy_number;
DROP INDEX IF EXISTS idx_insurance_policies_type;
DROP INDEX IF EXISTS idx_insurance_policies_status;
DROP INDEX IF EXISTS idx_insurance_policies_user_id;

DROP TABLE IF EXISTS insurance_claims;
DROP TABLE IF EXISTS insurance_policies; 