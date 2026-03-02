import { WIZARD_STEPS } from './types'
import { Check } from 'lucide-react'

interface Props {
  currentStep: number
  onStepClick: (step: number) => void
}

export default function WizardStepper({ currentStep, onStepClick }: Props) {
  return (
    <nav aria-label="Dealer creation wizard steps" className="flex items-center justify-between mb-8">
      {WIZARD_STEPS.map((step, index) => {
        const isCompleted = index < currentStep
        const isCurrent = index === currentStep

        return (
          <div key={step.label} className="flex items-center flex-1 last:flex-none">
            <button
              type="button"
              onClick={() => isCompleted ? onStepClick(index) : undefined}
              disabled={!isCompleted}
              className="flex items-center gap-3 group"
              aria-current={isCurrent ? 'step' : undefined}
              aria-label={`Step ${index + 1}: ${step.label}${isCompleted ? ' (completed)' : isCurrent ? ' (current)' : ''}`}
            >
              <div
                className={`flex items-center justify-center w-9 h-9 rounded-full text-sm font-medium transition-colors ${
                  isCompleted
                    ? 'bg-primary text-primary-foreground cursor-pointer'
                    : isCurrent
                    ? 'bg-primary text-primary-foreground ring-2 ring-primary/30 ring-offset-2 ring-offset-background'
                    : 'bg-muted text-muted-foreground'
                }`}
              >
                {isCompleted ? <Check className="h-4 w-4" /> : index + 1}
              </div>
              <div className="hidden sm:block">
                <p className={`text-sm font-medium ${isCurrent || isCompleted ? 'text-foreground' : 'text-muted-foreground'}`}>
                  {step.label}
                </p>
                <p className="text-xs text-muted-foreground">{step.description}</p>
              </div>
            </button>
            {index < WIZARD_STEPS.length - 1 && (
              <div className={`flex-1 h-px mx-4 ${isCompleted ? 'bg-primary' : 'bg-border'}`} />
            )}
          </div>
        )
      })}
    </nav>
  )
}
