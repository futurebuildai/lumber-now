-- name: GetInventoryItem :one
SELECT * FROM inventory WHERE id = $1;

-- name: GetInventoryBySKU :one
SELECT * FROM inventory WHERE dealer_id = $1 AND sku = $2;

-- name: ListInventory :many
SELECT * FROM inventory WHERE dealer_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3;

-- name: ListInventoryByCategory :many
SELECT * FROM inventory WHERE dealer_id = $1 AND category = $2 ORDER BY name ASC;

-- name: SearchInventory :many
SELECT * FROM inventory
WHERE dealer_id = $1
  AND to_tsvector('english', name || ' ' || description) @@ plainto_tsquery('english', $2)
ORDER BY name ASC
LIMIT $3;

-- name: CreateInventoryItem :one
INSERT INTO inventory (dealer_id, sku, name, description, category, unit, price, in_stock, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateInventoryItem :one
UPDATE inventory
SET name = $2, description = $3, category = $4, unit = $5, price = $6,
    in_stock = $7, metadata = $8, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpsertInventoryItem :one
INSERT INTO inventory (dealer_id, sku, name, description, category, unit, price, in_stock, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (dealer_id, sku) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    unit = EXCLUDED.unit,
    price = EXCLUDED.price,
    in_stock = EXCLUDED.in_stock,
    metadata = EXCLUDED.metadata,
    updated_at = now()
RETURNING *;

-- name: DeleteInventoryItem :exec
DELETE FROM inventory WHERE id = $1;

-- name: CountInventory :one
SELECT count(*) FROM inventory WHERE dealer_id = $1;
