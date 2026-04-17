-- ==========================================
-- Migration 000008: Auto-update updated_at trigger
-- ==========================================

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_society_updated BEFORE UPDATE ON society FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_flat_updated BEFORE UPDATE ON flat FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_user_updated BEFORE UPDATE ON app_user FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_complaint_updated BEFORE UPDATE ON complaint FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_payment_updated BEFORE UPDATE ON payment FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_vendor_updated BEFORE UPDATE ON vendor FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_listing_updated BEFORE UPDATE ON listing FOR EACH ROW EXECUTE FUNCTION update_updated_at();
