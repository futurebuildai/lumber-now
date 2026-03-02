import { describe, it, expect } from 'vitest'
import {
  required,
  isValidEmail,
  isValidPhone,
  isValidHexColor,
  isValidSlug,
  validateBasicInfo,
  validateBranding,
  validateContactInfo,
} from './validation'

describe('required', () => {
  it('returns error for empty string', () => {
    const err = required('', 'Name')
    expect(err).not.toBeNull()
    expect(err!.field).toBe('Name')
    expect(err!.message).toContain('required')
  })

  it('returns error for whitespace-only string', () => {
    expect(required('   ', 'Name')).not.toBeNull()
  })

  it('returns null for non-empty string', () => {
    expect(required('hello', 'Name')).toBeNull()
  })
})

describe('isValidEmail', () => {
  const valid = [
    'user@example.com',
    'user.name@example.co.uk',
    'user+tag@example.com',
    'first.last@sub.domain.com',
  ]
  const invalid = ['', 'not-email', '@domain.com', 'user@', 'user @domain.com', 'user@domain']

  valid.forEach((email) => {
    it(`accepts "${email}"`, () => {
      expect(isValidEmail(email)).toBe(true)
    })
  })

  invalid.forEach((email) => {
    it(`rejects "${email}"`, () => {
      expect(isValidEmail(email)).toBe(false)
    })
  })
})

describe('isValidPhone', () => {
  it('accepts 10-digit number', () => {
    expect(isValidPhone('5551234567')).toBe(true)
  })

  it('accepts formatted US number', () => {
    expect(isValidPhone('(555) 123-4567')).toBe(true)
  })

  it('accepts number with dashes', () => {
    expect(isValidPhone('555-123-4567')).toBe(true)
  })

  it('accepts 11-digit number with country code', () => {
    expect(isValidPhone('15551234567')).toBe(true)
  })

  it('rejects empty string', () => {
    expect(isValidPhone('')).toBe(false)
  })

  it('rejects too-short number', () => {
    expect(isValidPhone('12345')).toBe(false)
  })

  it('rejects too-long number', () => {
    expect(isValidPhone('123456789012')).toBe(false)
  })
})

describe('isValidHexColor', () => {
  it('accepts valid 6-char hex with hash', () => {
    expect(isValidHexColor('#1E40AF')).toBe(true)
    expect(isValidHexColor('#ff0000')).toBe(true)
    expect(isValidHexColor('#000000')).toBe(true)
    expect(isValidHexColor('#FFFFFF')).toBe(true)
  })

  it('rejects without hash', () => {
    expect(isValidHexColor('1E40AF')).toBe(false)
  })

  it('rejects 3-char shorthand', () => {
    expect(isValidHexColor('#F00')).toBe(false)
  })

  it('rejects invalid hex chars', () => {
    expect(isValidHexColor('#GGGGGG')).toBe(false)
  })

  it('rejects empty string', () => {
    expect(isValidHexColor('')).toBe(false)
  })
})

describe('isValidSlug', () => {
  it('accepts lowercase alphanumeric', () => {
    expect(isValidSlug('lumber-boss')).toBe(true)
    expect(isValidSlug('abc')).toBe(true)
    expect(isValidSlug('dealer123')).toBe(true)
  })

  it('rejects uppercase', () => {
    expect(isValidSlug('Lumber-Boss')).toBe(false)
  })

  it('rejects spaces', () => {
    expect(isValidSlug('lumber boss')).toBe(false)
  })

  it('rejects leading/trailing hyphens', () => {
    expect(isValidSlug('-lumber')).toBe(false)
    expect(isValidSlug('lumber-')).toBe(false)
  })

  it('rejects consecutive hyphens', () => {
    expect(isValidSlug('lumber--boss')).toBe(false)
  })

  it('rejects empty string', () => {
    expect(isValidSlug('')).toBe(false)
  })
})

describe('validateBasicInfo', () => {
  it('returns no errors for valid data', () => {
    const errors = validateBasicInfo({
      name: 'Lumber Boss',
      slug: 'lumber-boss',
      subdomain: 'lumberboss',
    })
    expect(errors).toHaveLength(0)
  })

  it('returns error for empty name', () => {
    const errors = validateBasicInfo({ name: '', slug: 'test', subdomain: 'test' })
    expect(errors.some((e) => e.field === 'name')).toBe(true)
  })

  it('returns error for name too short', () => {
    const errors = validateBasicInfo({ name: 'A', slug: 'test', subdomain: 'test' })
    expect(errors.some((e) => e.field === 'name' && e.message.includes('2 characters'))).toBe(true)
  })

  it('returns error for invalid slug', () => {
    const errors = validateBasicInfo({ name: 'Test', slug: 'INVALID SLUG', subdomain: 'test' })
    expect(errors.some((e) => e.field === 'slug')).toBe(true)
  })

  it('returns error for empty subdomain', () => {
    const errors = validateBasicInfo({ name: 'Test', slug: 'test', subdomain: '' })
    expect(errors.some((e) => e.field === 'subdomain')).toBe(true)
  })

  it('returns multiple errors for all-empty fields', () => {
    const errors = validateBasicInfo({ name: '', slug: '', subdomain: '' })
    expect(errors.length).toBeGreaterThanOrEqual(3)
  })
})

describe('validateBranding', () => {
  it('returns no errors for valid colors', () => {
    const errors = validateBranding({ primary_color: '#1E40AF', secondary_color: '#1E3A5F' })
    expect(errors).toHaveLength(0)
  })

  it('returns error for invalid primary color', () => {
    const errors = validateBranding({ primary_color: 'red', secondary_color: '#1E3A5F' })
    expect(errors.some((e) => e.field === 'primary_color')).toBe(true)
  })

  it('returns error for invalid secondary color', () => {
    const errors = validateBranding({ primary_color: '#1E40AF', secondary_color: 'not-a-color' })
    expect(errors.some((e) => e.field === 'secondary_color')).toBe(true)
  })
})

describe('validateContactInfo', () => {
  it('returns no errors for valid contact info', () => {
    const errors = validateContactInfo({ contact_email: 'info@test.com', contact_phone: '5551234567' })
    expect(errors).toHaveLength(0)
  })

  it('returns error for empty email', () => {
    const errors = validateContactInfo({ contact_email: '', contact_phone: '' })
    expect(errors.some((e) => e.field === 'contact_email')).toBe(true)
  })

  it('returns error for invalid email', () => {
    const errors = validateContactInfo({ contact_email: 'not-an-email', contact_phone: '' })
    expect(errors.some((e) => e.field === 'contact_email')).toBe(true)
  })

  it('allows empty phone (optional)', () => {
    const errors = validateContactInfo({ contact_email: 'info@test.com', contact_phone: '' })
    expect(errors).toHaveLength(0)
  })

  it('returns error for invalid phone', () => {
    const errors = validateContactInfo({ contact_email: 'info@test.com', contact_phone: '123' })
    expect(errors.some((e) => e.field === 'contact_phone')).toBe(true)
  })
})
