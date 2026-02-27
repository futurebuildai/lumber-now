import { useReducer } from 'react'
import type { WizardState, WizardAction, WizardFormData } from '../types'
import { INITIAL_FORM_DATA } from '../types'

const initialState: WizardState = {
  step: 0,
  formData: INITIAL_FORM_DATA,
  isSubmitting: false,
  createdDealerId: null,
}

function wizardReducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case 'SET_STEP':
      return { ...state, step: action.step }
    case 'UPDATE_FORM':
      return { ...state, formData: { ...state.formData, ...action.data } }
    case 'SUBMIT_START':
      return { ...state, isSubmitting: true }
    case 'SUBMIT_SUCCESS':
      return { ...state, isSubmitting: false, createdDealerId: action.dealerId }
    case 'SUBMIT_ERROR':
      return { ...state, isSubmitting: false }
    case 'RESET':
      return initialState
    default:
      return state
  }
}

export function useWizardState() {
  const [state, dispatch] = useReducer(wizardReducer, initialState)

  const nextStep = () => dispatch({ type: 'SET_STEP', step: state.step + 1 })
  const prevStep = () => dispatch({ type: 'SET_STEP', step: Math.max(0, state.step - 1) })
  const goToStep = (step: number) => dispatch({ type: 'SET_STEP', step })
  const updateForm = (data: Partial<WizardFormData>) => dispatch({ type: 'UPDATE_FORM', data })
  const submitStart = () => dispatch({ type: 'SUBMIT_START' })
  const submitSuccess = (dealerId: string) => dispatch({ type: 'SUBMIT_SUCCESS', dealerId })
  const submitError = () => dispatch({ type: 'SUBMIT_ERROR' })
  const reset = () => dispatch({ type: 'RESET' })

  return {
    state,
    nextStep,
    prevStep,
    goToStep,
    updateForm,
    submitStart,
    submitSuccess,
    submitError,
    reset,
  }
}
