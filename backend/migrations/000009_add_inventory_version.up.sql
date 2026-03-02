-- Add optimistic concurrency control to inventory
ALTER TABLE inventory ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Trigger to auto-increment inventory version on every update
CREATE OR REPLACE FUNCTION increment_inventory_version()
RETURNS TRIGGER AS $$
BEGIN
  NEW.version = OLD.version + 1;
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER inventory_version_increment
  BEFORE UPDATE ON inventory
  FOR EACH ROW
  EXECUTE FUNCTION increment_inventory_version();

-- Add CHECK constraints for data integrity
ALTER TABLE inventory ADD CONSTRAINT chk_inventory_price_nonneg CHECK (price >= 0);
ALTER TABLE inventory ADD CONSTRAINT chk_inventory_sku_nonempty CHECK (sku <> '');
ALTER TABLE inventory ADD CONSTRAINT chk_inventory_name_nonempty CHECK (name <> '');
