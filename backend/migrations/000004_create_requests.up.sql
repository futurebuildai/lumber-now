CREATE TYPE request_status AS ENUM (
    'pending',
    'processing',
    'parsed',
    'confirmed',
    'sent',
    'failed'
);

CREATE TYPE input_type AS ENUM (
    'text',
    'voice',
    'image',
    'pdf'
);

CREATE TABLE requests (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id         UUID NOT NULL REFERENCES dealers(id) ON DELETE CASCADE,
    contractor_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_rep_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    status            request_status NOT NULL DEFAULT 'pending',
    input_type        input_type NOT NULL DEFAULT 'text',
    raw_text          TEXT NOT NULL DEFAULT '',
    media_url         TEXT NOT NULL DEFAULT '',
    structured_items  JSONB NOT NULL DEFAULT '[]',
    ai_confidence     NUMERIC(5,4) NOT NULL DEFAULT 0,
    notes             TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_requests_dealer_id ON requests (dealer_id);
CREATE INDEX idx_requests_contractor ON requests (contractor_id);
CREATE INDEX idx_requests_assigned_rep ON requests (assigned_rep_id) WHERE assigned_rep_id IS NOT NULL;
CREATE INDEX idx_requests_status ON requests (dealer_id, status);
CREATE INDEX idx_requests_created_at ON requests (dealer_id, created_at DESC);
