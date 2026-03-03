import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useWizardState } from './hooks/useWizardState'
import { useCreateDealerMutation } from './hooks/useCreateDealer'
import { useGetDealer, useUpdateDealer } from '@/hooks/useDealers'
import WizardStepper from './WizardStepper'
import BasicInfoStep from './steps/BasicInfoStep'
import BrandingStep from './steps/BrandingStep'
import ContactInfoStep from './steps/ContactInfoStep'
import ReviewStep from './steps/ReviewStep'
import DealerPreviewPanel from './preview/DealerPreviewPanel'
import DealerCreatedSuccess from './success/DealerCreatedSuccess'
import { AlertCircle } from 'lucide-react'
import type { WizardFormData } from './types'

interface Props {
  mode?: 'create' | 'edit'
  dealerId?: string
}

export default function CreateDealerWizard({ mode = 'create', dealerId }: Props) {
  const { state, nextStep, prevStep, goToStep, updateForm, submitStart, submitSuccess, submitError, initEdit } = useWizardState()
  const createDealer = useCreateDealerMutation()
  const updateDealerMutation = useUpdateDealer()
  const qc = useQueryClient()
  const navigate = useNavigate()

  const isEdit = mode === 'edit'
  const { data: dealer, isLoading: isDealerLoading } = useGetDealer(dealerId ?? '')

  // Populate form when dealer data loads in edit mode
  useEffect(() => {
    if (isEdit && dealer && state.mode !== 'edit') {
      initEdit({
        name: dealer.name,
        slug: dealer.slug,
        subdomain: dealer.subdomain,
        logo_url: dealer.logo_url,
        logo_key: '',
        primary_color: dealer.primary_color,
        secondary_color: dealer.secondary_color,
        contact_email: dealer.contact_email,
        contact_phone: dealer.contact_phone,
        address: dealer.address,
      } satisfies WizardFormData)
    }
  }, [isEdit, dealer, state.mode, initEdit])

  const handleSubmit = async () => {
    submitStart()
    try {
      if (isEdit && dealerId) {
        await updateDealerMutation.mutateAsync({
          id: dealerId,
          name: state.formData.name,
          logo_url: state.formData.logo_url,
          primary_color: state.formData.primary_color,
          secondary_color: state.formData.secondary_color,
          contact_email: state.formData.contact_email,
          contact_phone: state.formData.contact_phone,
          address: state.formData.address,
        })
        navigate('/platform')
      } else {
        const result = await createDealer.mutateAsync(state.formData)
        qc.invalidateQueries({ queryKey: ['dealers'] })
        submitSuccess(result.id)
      }
    } catch {
      submitError()
    }
  }

  const mutationError = isEdit ? updateDealerMutation.isError : createDealer.isError

  // Loading state for edit mode
  if (isEdit && isDealerLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    )
  }

  // Success state (create mode only)
  if (!isEdit && state.createdDealerId) {
    return <DealerCreatedSuccess formData={state.formData} dealerId={state.createdDealerId} />
  }

  const showPreview = state.step >= 1

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">{isEdit ? 'Edit Dealer' : 'Create New Dealer'}</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {isEdit ? 'Update dealer branding and contact info.' : 'Set up a new dealer tenant with branding and contact info.'}
        </p>
      </div>

      <WizardStepper currentStep={state.step} onStepClick={goToStep} />

      {mutationError && (
        <div role="alert" aria-live="assertive" className="flex items-center gap-2 bg-destructive/10 text-destructive px-4 py-3 rounded-md text-sm">
          <AlertCircle className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
          {isEdit ? 'Failed to update dealer.' : 'Failed to create dealer. The slug or subdomain may already be taken.'}
        </div>
      )}

      <div className={`grid gap-8 ${showPreview ? 'grid-cols-[1fr_340px]' : ''}`}>
        {/* Step content */}
        <div className="bg-card border border-border rounded-lg p-6">
          {state.step === 0 && (
            <BasicInfoStep
              formData={state.formData}
              onUpdate={updateForm}
              onNext={nextStep}
              readOnly={isEdit}
            />
          )}
          {state.step === 1 && (
            <BrandingStep
              formData={state.formData}
              onUpdate={updateForm}
              onNext={nextStep}
              onBack={prevStep}
            />
          )}
          {state.step === 2 && (
            <ContactInfoStep
              formData={state.formData}
              onUpdate={updateForm}
              onNext={nextStep}
              onBack={prevStep}
            />
          )}
          {state.step === 3 && (
            <ReviewStep
              formData={state.formData}
              onBack={prevStep}
              onSubmit={handleSubmit}
              onEditStep={goToStep}
              isSubmitting={state.isSubmitting}
              mode={isEdit ? 'edit' : 'create'}
            />
          )}
        </div>

        {/* Live preview */}
        {showPreview && (
          <div className="bg-card border border-border rounded-lg p-6">
            <h3 className="text-sm font-semibold text-foreground mb-4 text-center">Live Preview</h3>
            <DealerPreviewPanel formData={state.formData} />
          </div>
        )}
      </div>
    </div>
  )
}
