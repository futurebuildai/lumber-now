ALTER TABLE inventory DROP CONSTRAINT IF EXISTS chk_inventory_name_nonempty;
ALTER TABLE inventory DROP CONSTRAINT IF EXISTS chk_inventory_sku_nonempty;
ALTER TABLE inventory DROP CONSTRAINT IF EXISTS chk_inventory_price_nonneg;
DROP TRIGGER IF EXISTS inventory_version_increment ON inventory;
DROP FUNCTION IF EXISTS increment_inventory_version();
ALTER TABLE inventory DROP COLUMN IF EXISTS version;
