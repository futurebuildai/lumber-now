CREATE TABLE inventory (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id   UUID NOT NULL REFERENCES dealers(id) ON DELETE CASCADE,
    sku         TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT '',
    unit        TEXT NOT NULL DEFAULT 'EA',
    price       NUMERIC(12,2) NOT NULL DEFAULT 0,
    in_stock    BOOLEAN NOT NULL DEFAULT true,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (dealer_id, sku)
);

CREATE INDEX idx_inventory_dealer_id ON inventory (dealer_id);
CREATE INDEX idx_inventory_sku ON inventory (dealer_id, sku);
CREATE INDEX idx_inventory_category ON inventory (dealer_id, category);
CREATE INDEX idx_inventory_name_search ON inventory USING gin (to_tsvector('english', name || ' ' || description));
