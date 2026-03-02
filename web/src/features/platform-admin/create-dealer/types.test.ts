import { describe, it, expect } from 'vitest'
import { WIZARD_STEPS, INITIAL_FORM_DATA } from './types'
import type { WizardFormData, WizardState, WizardAction } from './types'

describe('WIZARD_STEPS', () => {
  it('has exactly 4 steps', () => {
    expect(WIZARD_STEPS).toHaveLength(4)
  })

  it('each step has a label and description', () => {
    for (const step of WIZARD_STEPS) {
      expect(step.label).toBeTruthy()
      expect(step.description).toBeTruthy()
    }
  })

  it('steps are in correct order', () => {
    expect(WIZARD_STEPS[0].label).toBe('Basic Info')
    expect(WIZARD_STEPS[1].label).toBe('Branding')
    expect(WIZARD_STEPS[2].label).toBe('Contact')
    expect(WIZARD_STEPS[3].label).toBe('Review')
  })
})

describe('INITIAL_FORM_DATA', () => {
  it('has all required fields', () => {
    const requiredFields: (keyof WizardFormData)[] = [
      'name', 'slug', 'subdomain',
      'logo_url', 'logo_key', 'primary_color', 'secondary_color',
      'contact_email', 'contact_phone', 'address',
    ]
    for (const field of requiredFields) {
      expect(INITIAL_FORM_DATA).toHaveProperty(field)
    }
  })

  it('string fields default to empty string', () => {
    expect(INITIAL_FORM_DATA.name).toBe('')
    expect(INITIAL_FORM_DATA.slug).toBe('')
    expect(INITIAL_FORM_DATA.subdomain).toBe('')
    expect(INITIAL_FORM_DATA.contact_email).toBe('')
    expect(INITIAL_FORM_DATA.contact_phone).toBe('')
    expect(INITIAL_FORM_DATA.address).toBe('')
  })

  it('has default brand colors', () => {
    expect(INITIAL_FORM_DATA.primary_color).toMatch(/^#[0-9A-Fa-f]{6}$/)
    expect(INITIAL_FORM_DATA.secondary_color).toMatch(/^#[0-9A-Fa-f]{6}$/)
  })
})

describe('WizardState type shape', () => {
  it('can create a valid initial WizardState', () => {
    const state: WizardState = {
      step: 0,
      formData: INITIAL_FORM_DATA,
      isSubmitting: false,
      createdDealerId: null,
    }
    expect(state.step).toBe(0)
    expect(state.isSubmitting).toBe(false)
    expect(state.createdDealerId).toBeNull()
  })
})

describe('WizardAction type shape', () => {
  it('SET_STEP action has required fields', () => {
    const action: WizardAction = { type: 'SET_STEP', step: 2 }
    expect(action.type).toBe('SET_STEP')
  })

  it('UPDATE_FORM action accepts partial data', () => {
    const action: WizardAction = { type: 'UPDATE_FORM', data: { name: 'Test' } }
    expect(action.type).toBe('UPDATE_FORM')
  })

  it('SUBMIT_SUCCESS action has dealerId', () => {
    const action: WizardAction = { type: 'SUBMIT_SUCCESS', dealerId: 'abc' }
    expect(action.type).toBe('SUBMIT_SUCCESS')
  })

  it('simple actions have no payload', () => {
    const actions: WizardAction[] = [
      { type: 'SUBMIT_START' },
      { type: 'SUBMIT_ERROR' },
      { type: 'RESET' },
    ]
    expect(actions).toHaveLength(3)
  })
})
