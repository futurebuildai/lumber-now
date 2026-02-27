export interface WizardFormData {
  // Step 1: Basic Info
  name: string
  slug: string
  subdomain: string

  // Step 2: Branding
  logo_url: string
  logo_key: string
  primary_color: string
  secondary_color: string

  // Step 3: Contact Info
  contact_email: string
  contact_phone: string
  address: string
}

export interface WizardState {
  step: number
  formData: WizardFormData
  isSubmitting: boolean
  createdDealerId: string | null
}

export type WizardAction =
  | { type: 'SET_STEP'; step: number }
  | { type: 'UPDATE_FORM'; data: Partial<WizardFormData> }
  | { type: 'SUBMIT_START' }
  | { type: 'SUBMIT_SUCCESS'; dealerId: string }
  | { type: 'SUBMIT_ERROR' }
  | { type: 'RESET' }

export const WIZARD_STEPS = [
  { label: 'Basic Info', description: 'Name and slug' },
  { label: 'Branding', description: 'Logo and colors' },
  { label: 'Contact', description: 'Email and phone' },
  { label: 'Review', description: 'Confirm details' },
] as const

export const INITIAL_FORM_DATA: WizardFormData = {
  name: '',
  slug: '',
  subdomain: '',
  logo_url: '',
  logo_key: '',
  primary_color: '#1E40AF',
  secondary_color: '#1E3A5F',
  contact_email: '',
  contact_phone: '',
  address: '',
}
