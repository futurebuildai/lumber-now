import { describe, it, expect } from 'vitest'
import type {
  Role,
  RequestStatus,
  InputType,
  Dealer,
  User,
  InventoryItem,
  StructuredItem,
  MaterialRequest,
  TenantConfig,
  TokenPair,
} from './index'

describe('Role type', () => {
  it('allows valid role values', () => {
    const roles: Role[] = ['platform_admin', 'dealer_admin', 'sales_rep', 'contractor']
    expect(roles).toHaveLength(4)
    expect(new Set(roles).size).toBe(4)
  })
})

describe('RequestStatus type', () => {
  it('allows all 6 status values', () => {
    const statuses: RequestStatus[] = ['pending', 'processing', 'parsed', 'confirmed', 'sent', 'failed']
    expect(statuses).toHaveLength(6)
    expect(new Set(statuses).size).toBe(6)
  })
})

describe('InputType type', () => {
  it('allows all 4 input types', () => {
    const types: InputType[] = ['text', 'voice', 'image', 'pdf']
    expect(types).toHaveLength(4)
    expect(new Set(types).size).toBe(4)
  })
})

describe('Dealer interface', () => {
  it('has the expected shape', () => {
    const dealer: Dealer = {
      id: 'test-id',
      name: 'Test Dealer',
      slug: 'test-dealer',
      subdomain: 'test',
      logo_url: '',
      primary_color: '#1E40AF',
      secondary_color: '#1E3A5F',
      contact_email: 'test@example.com',
      contact_phone: '555-1234',
      address: '123 Main St',
      active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    expect(dealer.id).toBe('test-id')
    expect(dealer.active).toBe(true)
  })
})

describe('User interface', () => {
  it('has the expected shape', () => {
    const user: User = {
      id: 'user-id',
      dealer_id: 'dealer-id',
      email: 'user@example.com',
      full_name: 'Test User',
      phone: '555-0000',
      role: 'contractor',
      active: true,
      created_at: '2024-01-01T00:00:00Z',
    }
    expect(user.role).toBe('contractor')
    expect(user.assigned_rep_id).toBeUndefined()
  })

  it('allows optional assigned_rep_id', () => {
    const user: User = {
      id: 'user-id',
      dealer_id: 'dealer-id',
      email: 'user@example.com',
      full_name: 'Test User',
      phone: '555-0000',
      role: 'contractor',
      assigned_rep_id: 'rep-id',
      active: true,
      created_at: '2024-01-01T00:00:00Z',
    }
    expect(user.assigned_rep_id).toBe('rep-id')
  })
})

describe('StructuredItem interface', () => {
  it('has the expected shape', () => {
    const item: StructuredItem = {
      sku: 'LBR-2X4',
      name: '2x4 Lumber',
      quantity: 100,
      unit: 'pieces',
      confidence: 0.95,
      matched: true,
    }
    expect(item.sku).toBe('LBR-2X4')
    expect(item.notes).toBeUndefined()
  })

  it('allows optional notes', () => {
    const item: StructuredItem = {
      sku: 'LBR-2X4',
      name: '2x4 Lumber',
      quantity: 100,
      unit: 'pieces',
      confidence: 0.95,
      matched: true,
      notes: 'rush order',
    }
    expect(item.notes).toBe('rush order')
  })
})

describe('MaterialRequest interface', () => {
  it('has the expected shape', () => {
    const request: MaterialRequest = {
      id: 'req-id',
      dealer_id: 'dealer-id',
      contractor_id: 'contractor-id',
      status: 'pending',
      input_type: 'text',
      raw_text: '100 2x4 boards',
      media_url: '',
      structured_items: [],
      ai_confidence: '0.0000',
      notes: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    expect(request.status).toBe('pending')
    expect(request.structured_items).toHaveLength(0)
  })
})

describe('TenantConfig interface', () => {
  it('has the expected shape', () => {
    const config: TenantConfig = {
      dealer_id: 'dealer-id',
      name: 'Test',
      slug: 'test',
      logo_url: '',
      primary_color: '#000000',
      secondary_color: '#111111',
      contact_email: 'test@test.com',
      contact_phone: '555-1234',
    }
    expect(config.dealer_id).toBe('dealer-id')
  })
})

describe('TokenPair interface', () => {
  it('has the expected shape', () => {
    const tokens: TokenPair = {
      access_token: 'eyJhbGciOiJIUzI1NiJ9...',
      refresh_token: 'eyJhbGciOiJIUzI1NiJ9...',
      expires_at: 1700000000,
    }
    expect(tokens.access_token).toBeTruthy()
    expect(tokens.expires_at).toBeGreaterThan(0)
  })

  it('expires_at is a numeric timestamp', () => {
    const tokens: TokenPair = {
      access_token: 'at',
      refresh_token: 'rt',
      expires_at: Date.now() / 1000 + 3600,
    }
    expect(Number.isFinite(tokens.expires_at)).toBe(true)
    expect(tokens.expires_at).toBeGreaterThan(0)
  })
})

// ---------------------------------------------------------------------------
// Additional Dealer tests
// ---------------------------------------------------------------------------

describe('Dealer interface - additional', () => {
  it('string fields are strings', () => {
    const dealer: Dealer = {
      id: '550e8400-e29b-41d4-a716-446655440000',
      name: 'Acme Lumber',
      slug: 'acme-lumber',
      subdomain: 'acme',
      logo_url: 'https://cdn.example.com/logo.png',
      primary_color: '#1a73e8',
      secondary_color: '#ffffff',
      contact_email: 'info@acmelumber.com',
      contact_phone: '+15551234567',
      address: '123 Main St, Springfield, IL',
      active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-06-15T12:00:00Z',
    }
    expect(typeof dealer.id).toBe('string')
    expect(typeof dealer.name).toBe('string')
    expect(typeof dealer.slug).toBe('string')
    expect(typeof dealer.subdomain).toBe('string')
    expect(typeof dealer.contact_email).toBe('string')
    expect(typeof dealer.contact_phone).toBe('string')
    expect(typeof dealer.address).toBe('string')
    expect(typeof dealer.active).toBe('boolean')
  })

  it('has all 13 required fields', () => {
    const dealer: Dealer = {
      id: 'd',
      name: 'n',
      slug: 's',
      subdomain: 'sub',
      logo_url: '',
      primary_color: '#000',
      secondary_color: '#fff',
      contact_email: 'e@e.com',
      contact_phone: '555',
      address: 'addr',
      active: false,
      created_at: '',
      updated_at: '',
    }
    const keys = Object.keys(dealer)
    expect(keys).toHaveLength(13)
  })
})

// ---------------------------------------------------------------------------
// Additional InventoryItem tests
// ---------------------------------------------------------------------------

describe('InventoryItem interface', () => {
  const item: InventoryItem = {
    id: 'inv-001',
    dealer_id: 'dealer-001',
    sku: 'LBR-2X4-8',
    name: '2x4 Lumber 8ft',
    description: 'Standard grade 2x4 lumber, 8 foot length',
    category: 'lumber',
    unit: 'pieces',
    price: '5.99',
    in_stock: true,
    metadata: { grade: 'standard', species: 'pine' },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-06-01T00:00:00Z',
  }

  it('has all required fields', () => {
    expect(item.sku).toBe('LBR-2X4-8')
    expect(item.price).toBe('5.99')
    expect(typeof item.price).toBe('string')
    expect(item.in_stock).toBe(true)
  })

  it('metadata is an object', () => {
    expect(typeof item.metadata).toBe('object')
    expect(item.metadata).not.toBeNull()
  })

  it('metadata can be empty', () => {
    const emptyMeta: InventoryItem = { ...item, metadata: {} }
    expect(Object.keys(emptyMeta.metadata)).toHaveLength(0)
  })

  it('has 12 required fields', () => {
    expect(Object.keys(item)).toHaveLength(12)
  })
})

// ---------------------------------------------------------------------------
// Additional StructuredItem tests
// ---------------------------------------------------------------------------

describe('StructuredItem interface - additional', () => {
  it('confidence is a number', () => {
    const item: StructuredItem = {
      sku: 'B',
      name: 'Item B',
      quantity: 5,
      unit: 'ft',
      confidence: 0.72,
      matched: true,
    }
    expect(typeof item.confidence).toBe('number')
    expect(item.confidence).toBeGreaterThanOrEqual(0)
    expect(item.confidence).toBeLessThanOrEqual(1)
  })

  it('quantity is a number', () => {
    const item: StructuredItem = {
      sku: 'C',
      name: 'Item C',
      quantity: 42.5,
      unit: 'lbs',
      confidence: 0.99,
      matched: false,
    }
    expect(typeof item.quantity).toBe('number')
    expect(item.quantity).toBe(42.5)
  })
})

// ---------------------------------------------------------------------------
// Additional MaterialRequest tests
// ---------------------------------------------------------------------------

describe('MaterialRequest interface - additional', () => {
  it('accepts request with assigned_rep_id', () => {
    const request: MaterialRequest = {
      id: 'req-002',
      dealer_id: 'dealer-001',
      contractor_id: 'contractor-001',
      assigned_rep_id: 'rep-001',
      status: 'parsed',
      input_type: 'image',
      raw_text: '',
      media_url: 'https://cdn.example.com/photo.jpg',
      structured_items: [
        {
          sku: 'LBR-2X4-8',
          name: '2x4 Lumber 8ft',
          quantity: 50,
          unit: 'pieces',
          confidence: 0.92,
          matched: true,
        },
      ],
      ai_confidence: '0.9200',
      notes: 'Parsed from photo',
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T12:00:00Z',
    }
    expect(request.assigned_rep_id).toBe('rep-001')
    expect(request.structured_items).toHaveLength(1)
    expect(request.structured_items[0].sku).toBe('LBR-2X4-8')
  })

  it('structured_items can be an empty array', () => {
    const request: MaterialRequest = {
      id: 'req-003',
      dealer_id: 'dealer-001',
      contractor_id: 'contractor-002',
      status: 'failed',
      input_type: 'voice',
      raw_text: '',
      media_url: 'https://cdn.example.com/voice.m4a',
      structured_items: [],
      ai_confidence: '0.0000',
      notes: 'Transcription failed',
      created_at: '2024-03-01T00:00:00Z',
      updated_at: '2024-03-01T00:05:00Z',
    }
    expect(request.structured_items).toHaveLength(0)
    expect(Array.isArray(request.structured_items)).toBe(true)
  })
})

// ---------------------------------------------------------------------------
// Cross-type consistency checks
// ---------------------------------------------------------------------------

describe('Cross-type consistency', () => {
  it('MaterialRequest.status is a valid RequestStatus', () => {
    const validStatuses: RequestStatus[] = [
      'pending',
      'processing',
      'parsed',
      'confirmed',
      'sent',
      'failed',
    ]
    const request: MaterialRequest = {
      id: 'req-check',
      dealer_id: 'd',
      contractor_id: 'c',
      status: 'confirmed',
      input_type: 'text',
      raw_text: 'test',
      media_url: '',
      structured_items: [],
      ai_confidence: '0.5',
      notes: '',
      created_at: '',
      updated_at: '',
    }
    expect(validStatuses).toContain(request.status)
  })

  it('MaterialRequest.input_type is a valid InputType', () => {
    const validTypes: InputType[] = ['text', 'voice', 'image', 'pdf']
    const request: MaterialRequest = {
      id: 'req-check2',
      dealer_id: 'd',
      contractor_id: 'c',
      status: 'pending',
      input_type: 'pdf',
      raw_text: '',
      media_url: 'https://cdn.example.com/doc.pdf',
      structured_items: [],
      ai_confidence: '0.0',
      notes: '',
      created_at: '',
      updated_at: '',
    }
    expect(validTypes).toContain(request.input_type)
  })

  it('User.role is a valid Role', () => {
    const validRoles: Role[] = ['platform_admin', 'dealer_admin', 'sales_rep', 'contractor']
    const user: User = {
      id: 'u1',
      dealer_id: 'd1',
      email: 'test@test.com',
      full_name: 'Test User',
      phone: '555-0000',
      role: 'dealer_admin',
      active: true,
      created_at: '',
    }
    expect(validRoles).toContain(user.role)
  })
})
