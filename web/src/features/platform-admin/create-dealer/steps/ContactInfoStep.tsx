import type { WizardFormData } from '../types'
import { Mail, Phone, MapPin } from 'lucide-react'

interface Props {
  formData: WizardFormData
  onUpdate: (data: Partial<WizardFormData>) => void
  onNext: () => void
  onBack: () => void
}

export default function ContactInfoStep({ formData, onUpdate, onNext, onBack }: Props) {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-foreground">Contact Information</h2>
        <p className="text-sm text-muted-foreground mt-1">How customers will reach this dealer.</p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground">Contact Email</label>
          <div className="relative">
            <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="email"
              value={formData.contact_email}
              onChange={(e) => onUpdate({ contact_email: e.target.value })}
              placeholder="orders@acmelumber.com"
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground">Contact Phone</label>
          <div className="relative">
            <Phone className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="tel"
              value={formData.contact_phone}
              onChange={(e) => onUpdate({ contact_phone: e.target.value })}
              placeholder="(555) 123-4567"
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground">Address</label>
          <div className="relative">
            <MapPin className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
            <textarea
              value={formData.address}
              onChange={(e) => onUpdate({ address: e.target.value })}
              rows={3}
              placeholder="123 Main St, Springfield, IL 62701"
              className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
          </div>
        </div>
      </div>

      <div className="flex justify-between pt-4">
        <button
          onClick={onBack}
          className="inline-flex items-center rounded-md border border-input bg-background px-6 py-2.5 text-sm font-medium text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        >
          Back
        </button>
        <button
          onClick={onNext}
          className="inline-flex items-center rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  )
}
