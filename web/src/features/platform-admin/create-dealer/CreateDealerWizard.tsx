import { useQueryClient } from '@tanstack/react-query'
import { useWizardState } from './hooks/useWizardState'
import { useCreateDealerMutation } from './hooks/useCreateDealer'
import WizardStepper from './WizardStepper'
import BasicInfoStep from './steps/BasicInfoStep'
import BrandingStep from './steps/BrandingStep'
import ContactInfoStep from './steps/ContactInfoStep'
import ReviewStep from './steps/ReviewStep'
import DealerPreviewPanel from './preview/DealerPreviewPanel'
import DealerCreatedSuccess from './success/DealerCreatedSuccess'
import { AlertCircle } from 'lucide-react'

export default function CreateDealerWizard() {
  const { state, nextStep, prevStep, goToStep, updateForm, submitStart, submitSuccess, submitError } = useWizardState()
  const createDealer = useCreateDealerMutation()
  const qc = useQueryClient()

  const handleSubmit = async () => {
    submitStart()
    try {
      const dealer = await createDealer.mutateAsync(state.formData)
      qc.invalidateQueries({ queryKey: ['dealers'] })
      submitSuccess(dealer.id)
    } catch {
      submitError()
    }
  }

  // Success state
  if (state.createdDealerId) {
    return <DealerCreatedSuccess formData={state.formData} dealerId={state.createdDealerId} />
  }

  const showPreview = state.step >= 1

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Create New Dealer</h1>
        <p className="text-sm text-muted-foreground mt-1">Set up a new dealer tenant with branding and contact info.</p>
      </div>

      <WizardStepper currentStep={state.step} onStepClick={goToStep} />

      {createDealer.isError && (
        <div role="alert" aria-live="assertive" className="flex items-center gap-2 bg-destructive/10 text-destructive px-4 py-3 rounded-md text-sm">
          <AlertCircle className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
          Failed to create dealer. The slug or subdomain may already be taken.
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
