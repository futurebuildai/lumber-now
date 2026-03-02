import type { WizardFormData } from '@/features/platform-admin/create-dealer/types'

export interface ValidationError {
  field: string
  message: string
}

/** Validate a non-empty required string field. */
export function required(value: string, fieldName: string): ValidationError | null {
  if (!value || value.trim().length === 0) {
    return { field: fieldName, message: `${fieldName} is required` }
  }
  return null
}

/** Validate an email address format. */
export function isValidEmail(email: string): boolean {
  if (!email) return false
  // RFC 5322-ish pattern (not perfect, but catches common mistakes)
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

/** Validate a US phone number (10 digits, optional formatting). */
export function isValidPhone(phone: string): boolean {
  if (!phone) return false
  const digits = phone.replace(/\D/g, '')
  return digits.length === 10 || digits.length === 11
}

/** Validate a hex color code. */
export function isValidHexColor(color: string): boolean {
  return /^#[0-9A-Fa-f]{6}$/.test(color)
}

/** Validate a slug (lowercase, alphanumeric, hyphens). */
export function isValidSlug(slug: string): boolean {
  return /^[a-z0-9]+(-[a-z0-9]+)*$/.test(slug)
}

/** Validate Step 1 (Basic Info) of the dealer wizard. */
export function validateBasicInfo(data: Pick<WizardFormData, 'name' | 'slug' | 'subdomain'>): ValidationError[] {
  const errors: ValidationError[] = []

  const nameErr = required(data.name, 'name')
  if (nameErr) errors.push(nameErr)
  else if (data.name.length < 2) errors.push({ field: 'name', message: 'Name must be at least 2 characters' })
  else if (data.name.length > 100) errors.push({ field: 'name', message: 'Name must be at most 100 characters' })

  const slugErr = required(data.slug, 'slug')
  if (slugErr) errors.push(slugErr)
  else if (!isValidSlug(data.slug)) errors.push({ field: 'slug', message: 'Slug must be lowercase alphanumeric with hyphens' })

  const subErr = required(data.subdomain, 'subdomain')
  if (subErr) errors.push(subErr)
  else if (!isValidSlug(data.subdomain)) errors.push({ field: 'subdomain', message: 'Subdomain must be lowercase alphanumeric with hyphens' })

  return errors
}

/** Validate Step 2 (Branding) of the dealer wizard. */
export function validateBranding(data: Pick<WizardFormData, 'primary_color' | 'secondary_color'>): ValidationError[] {
  const errors: ValidationError[] = []

  if (!isValidHexColor(data.primary_color)) {
    errors.push({ field: 'primary_color', message: 'Primary color must be a valid hex color (#RRGGBB)' })
  }
  if (!isValidHexColor(data.secondary_color)) {
    errors.push({ field: 'secondary_color', message: 'Secondary color must be a valid hex color (#RRGGBB)' })
  }

  return errors
}

/** Validate Step 3 (Contact Info) of the dealer wizard. */
export function validateContactInfo(data: Pick<WizardFormData, 'contact_email' | 'contact_phone'>): ValidationError[] {
  const errors: ValidationError[] = []

  const emailErr = required(data.contact_email, 'contact_email')
  if (emailErr) errors.push(emailErr)
  else if (!isValidEmail(data.contact_email)) errors.push({ field: 'contact_email', message: 'Invalid email address' })

  if (data.contact_phone && !isValidPhone(data.contact_phone)) {
    errors.push({ field: 'contact_phone', message: 'Phone must be 10 or 11 digits' })
  }

  return errors
}
