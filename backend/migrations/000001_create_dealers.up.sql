CREATE TABLE dealers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    subdomain   TEXT NOT NULL UNIQUE,
    logo_url    TEXT NOT NULL DEFAULT '',
    primary_color   TEXT NOT NULL DEFAULT '#1E40AF',
    secondary_color TEXT NOT NULL DEFAULT '#1E3A5F',
    contact_email   TEXT NOT NULL DEFAULT '',
    contact_phone   TEXT NOT NULL DEFAULT '',
    address         TEXT NOT NULL DEFAULT '',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_dealers_slug ON dealers (slug);
CREATE INDEX idx_dealers_subdomain ON dealers (subdomain);
CREATE INDEX idx_dealers_active ON dealers (active) WHERE active = true;
