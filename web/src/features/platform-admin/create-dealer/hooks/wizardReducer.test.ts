import { describe, it, expect } from 'vitest'
import { wizardReducer } from './useWizardState'
import { INITIAL_FORM_DATA } from '../types'
import type { WizardState, WizardAction } from '../types'

const initialState: WizardState = {
  step: 0,
  formData: INITIAL_FORM_DATA,
  isSubmitting: false,
  createdDealerId: null,
}

describe('wizardReducer', () => {
  describe('SET_STEP', () => {
    it('sets step to the given value', () => {
      const result = wizardReducer(initialState, { type: 'SET_STEP', step: 2 })
      expect(result.step).toBe(2)
    })

    it('preserves other state when setting step', () => {
      const state: WizardState = {
        ...initialState,
        formData: { ...INITIAL_FORM_DATA, name: 'Test Dealer' },
        isSubmitting: false,
      }
      const result = wizardReducer(state, { type: 'SET_STEP', step: 3 })
      expect(result.step).toBe(3)
      expect(result.formData.name).toBe('Test Dealer')
    })

    it('allows setting step to 0', () => {
      const state = { ...initialState, step: 3 }
      const result = wizardReducer(state, { type: 'SET_STEP', step: 0 })
      expect(result.step).toBe(0)
    })
  })

  describe('UPDATE_FORM', () => {
    it('updates a single form field', () => {
      const result = wizardReducer(initialState, {
        type: 'UPDATE_FORM',
        data: { name: 'Lumber Boss' },
      })
      expect(result.formData.name).toBe('Lumber Boss')
    })

    it('updates multiple form fields at once', () => {
      const result = wizardReducer(initialState, {
        type: 'UPDATE_FORM',
        data: {
          name: 'Lumber Boss',
          slug: 'lumber-boss',
          contact_email: 'info@lumberboss.com',
        },
      })
      expect(result.formData.name).toBe('Lumber Boss')
      expect(result.formData.slug).toBe('lumber-boss')
      expect(result.formData.contact_email).toBe('info@lumberboss.com')
    })

    it('preserves fields not in the update', () => {
      const state: WizardState = {
        ...initialState,
        formData: { ...INITIAL_FORM_DATA, name: 'Existing Name' },
      }
      const result = wizardReducer(state, {
        type: 'UPDATE_FORM',
        data: { slug: 'new-slug' },
      })
      expect(result.formData.name).toBe('Existing Name')
      expect(result.formData.slug).toBe('new-slug')
    })

    it('does not change step when updating form', () => {
      const state = { ...initialState, step: 2 }
      const result = wizardReducer(state, {
        type: 'UPDATE_FORM',
        data: { name: 'Test' },
      })
      expect(result.step).toBe(2)
    })

    it('overwrites default colors', () => {
      expect(initialState.formData.primary_color).toBe('#1E40AF')
      const result = wizardReducer(initialState, {
        type: 'UPDATE_FORM',
        data: { primary_color: '#FF0000' },
      })
      expect(result.formData.primary_color).toBe('#FF0000')
    })
  })

  describe('SUBMIT_START', () => {
    it('sets isSubmitting to true', () => {
      const result = wizardReducer(initialState, { type: 'SUBMIT_START' })
      expect(result.isSubmitting).toBe(true)
    })

    it('preserves form data', () => {
      const state: WizardState = {
        ...initialState,
        formData: { ...INITIAL_FORM_DATA, name: 'Test' },
      }
      const result = wizardReducer(state, { type: 'SUBMIT_START' })
      expect(result.formData.name).toBe('Test')
      expect(result.isSubmitting).toBe(true)
    })
  })

  describe('SUBMIT_SUCCESS', () => {
    it('sets isSubmitting to false and stores dealer ID', () => {
      const state: WizardState = { ...initialState, isSubmitting: true }
      const result = wizardReducer(state, {
        type: 'SUBMIT_SUCCESS',
        dealerId: 'abc-123',
      })
      expect(result.isSubmitting).toBe(false)
      expect(result.createdDealerId).toBe('abc-123')
    })
  })

  describe('SUBMIT_ERROR', () => {
    it('sets isSubmitting to false', () => {
      const state: WizardState = { ...initialState, isSubmitting: true }
      const result = wizardReducer(state, { type: 'SUBMIT_ERROR' })
      expect(result.isSubmitting).toBe(false)
    })

    it('does not clear createdDealerId', () => {
      const state: WizardState = {
        ...initialState,
        isSubmitting: true,
        createdDealerId: 'existing-id',
      }
      const result = wizardReducer(state, { type: 'SUBMIT_ERROR' })
      expect(result.createdDealerId).toBe('existing-id')
    })
  })

  describe('RESET', () => {
    it('returns to initial state', () => {
      const state: WizardState = {
        step: 3,
        formData: {
          ...INITIAL_FORM_DATA,
          name: 'Modified',
          slug: 'modified',
        },
        isSubmitting: true,
        createdDealerId: 'some-id',
      }
      const result = wizardReducer(state, { type: 'RESET' })
      expect(result.step).toBe(0)
      expect(result.formData).toEqual(INITIAL_FORM_DATA)
      expect(result.isSubmitting).toBe(false)
      expect(result.createdDealerId).toBeNull()
    })
  })

  describe('unknown action', () => {
    it('returns state unchanged for unknown action type', () => {
      const result = wizardReducer(initialState, { type: 'UNKNOWN' } as unknown as WizardAction)
      expect(result).toEqual(initialState)
    })
  })

  describe('full wizard flow', () => {
    it('simulates a complete wizard lifecycle', () => {
      // Step 1: Fill basic info
      let state = wizardReducer(initialState, {
        type: 'UPDATE_FORM',
        data: { name: 'Lumber Boss', slug: 'lumber-boss', subdomain: 'lumberboss' },
      })
      state = wizardReducer(state, { type: 'SET_STEP', step: 1 })
      expect(state.step).toBe(1)

      // Step 2: Branding
      state = wizardReducer(state, {
        type: 'UPDATE_FORM',
        data: { primary_color: '#FF0000', secondary_color: '#00FF00' },
      })
      state = wizardReducer(state, { type: 'SET_STEP', step: 2 })

      // Step 3: Contact
      state = wizardReducer(state, {
        type: 'UPDATE_FORM',
        data: { contact_email: 'info@lumberboss.com', contact_phone: '555-1234' },
      })
      state = wizardReducer(state, { type: 'SET_STEP', step: 3 })

      // Step 4: Submit
      state = wizardReducer(state, { type: 'SUBMIT_START' })
      expect(state.isSubmitting).toBe(true)

      state = wizardReducer(state, { type: 'SUBMIT_SUCCESS', dealerId: 'new-dealer-uuid' })
      expect(state.isSubmitting).toBe(false)
      expect(state.createdDealerId).toBe('new-dealer-uuid')
      expect(state.formData.name).toBe('Lumber Boss')
      expect(state.formData.contact_email).toBe('info@lumberboss.com')

      // Reset for another wizard run
      state = wizardReducer(state, { type: 'RESET' })
      expect(state).toEqual(initialState)
    })

    it('handles submit error and retry', () => {
      let state = wizardReducer(initialState, {
        type: 'UPDATE_FORM',
        data: { name: 'Test Dealer' },
      })
      state = wizardReducer(state, { type: 'SET_STEP', step: 3 })
      state = wizardReducer(state, { type: 'SUBMIT_START' })
      expect(state.isSubmitting).toBe(true)

      // Error
      state = wizardReducer(state, { type: 'SUBMIT_ERROR' })
      expect(state.isSubmitting).toBe(false)
      expect(state.formData.name).toBe('Test Dealer') // Data preserved

      // Retry
      state = wizardReducer(state, { type: 'SUBMIT_START' })
      state = wizardReducer(state, { type: 'SUBMIT_SUCCESS', dealerId: 'retry-uuid' })
      expect(state.createdDealerId).toBe('retry-uuid')
    })
  })

  describe('INITIAL_FORM_DATA', () => {
    it('has expected default values', () => {
      expect(INITIAL_FORM_DATA.name).toBe('')
      expect(INITIAL_FORM_DATA.slug).toBe('')
      expect(INITIAL_FORM_DATA.subdomain).toBe('')
      expect(INITIAL_FORM_DATA.logo_url).toBe('')
      expect(INITIAL_FORM_DATA.logo_key).toBe('')
      expect(INITIAL_FORM_DATA.primary_color).toBe('#1E40AF')
      expect(INITIAL_FORM_DATA.secondary_color).toBe('#1E3A5F')
      expect(INITIAL_FORM_DATA.contact_email).toBe('')
      expect(INITIAL_FORM_DATA.contact_phone).toBe('')
      expect(INITIAL_FORM_DATA.address).toBe('')
    })
  })
})
