CREATE TYPE user_role AS ENUM (
    'platform_admin',
    'dealer_admin',
    'sales_rep',
    'contractor'
);

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id       UUID NOT NULL REFERENCES dealers(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    full_name       TEXT NOT NULL,
    phone           TEXT NOT NULL DEFAULT '',
    role            user_role NOT NULL DEFAULT 'contractor',
    assigned_rep_id UUID REFERENCES users(id) ON DELETE SET NULL,
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (dealer_id, email)
);

CREATE INDEX idx_users_dealer_id ON users (dealer_id);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (dealer_id, role);
CREATE INDEX idx_users_assigned_rep ON users (assigned_rep_id) WHERE assigned_rep_id IS NOT NULL;
