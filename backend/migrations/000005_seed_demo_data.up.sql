-- Seed demo dealer: Lumber Boss
INSERT INTO dealers (id, name, slug, subdomain, contact_email, contact_phone, address, active)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'Lumber Boss',
    'lumber-boss',
    'lumberboss',
    'orders@lumberboss.com',
    '555-867-5309',
    '123 Timber Lane, Portland, OR 97201',
    true
) ON CONFLICT (id) DO NOTHING;

-- Seed platform admin: colton@futurebuild.ai
-- Password: LumberNow2024!
INSERT INTO users (id, dealer_id, email, password_hash, full_name, role, active)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'colton@futurebuild.ai',
    '$2b$10$QxLe64oGn951PIdD0gjPBODLsvD/k//dRw8EhiG2oq.Fi.Zq7Htg2',
    'Colton Futurebuild',
    'platform_admin',
    true
) ON CONFLICT (id) DO NOTHING;
