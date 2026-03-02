import type { WizardFormData } from '../types'
import { Building2, Palette, Mail, Phone, MapPin, Globe, Loader2 } from 'lucide-react'

interface Props {
  formData: WizardFormData
  onBack: () => void
  onSubmit: () => void
  onEditStep: (step: number) => void
  isSubmitting: boolean
}

function ReviewRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between py-2">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-medium text-foreground text-right max-w-[60%] truncate">{value || '-'}</span>
    </div>
  )
}

export default function ReviewStep({ formData, onBack, onSubmit, onEditStep, isSubmitting }: Props) {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-foreground">Review & Create</h2>
        <p className="text-sm text-muted-foreground mt-1">Verify the dealer details before creating.</p>
      </div>

      <div className="space-y-4">
        {/* Basic Info */}
        <div className="bg-muted/30 rounded-lg border border-border p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Building2 className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold text-foreground">Basic Info</h3>
            </div>
            <button onClick={() => onEditStep(0)} className="text-xs font-medium text-primary hover:text-primary/80" aria-label="Edit basic info">
              Edit
            </button>
          </div>
          <div className="divide-y divide-border">
            <ReviewRow label="Name" value={formData.name} />
            <ReviewRow label="Slug" value={formData.slug} />
            <ReviewRow label="Subdomain" value={formData.subdomain} />
          </div>
        </div>

        {/* Branding */}
        <div className="bg-muted/30 rounded-lg border border-border p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Palette className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold text-foreground">Branding</h3>
            </div>
            <button onClick={() => onEditStep(1)} className="text-xs font-medium text-primary hover:text-primary/80" aria-label="Edit branding">
              Edit
            </button>
          </div>
          <div className="divide-y divide-border">
            {formData.logo_url && (
              <div className="flex justify-between items-center py-2">
                <span className="text-sm text-muted-foreground">Logo</span>
                <img src={formData.logo_url} alt="" className="h-8 w-8 object-contain rounded" />
              </div>
            )}
            <div className="flex justify-between items-center py-2">
              <span className="text-sm text-muted-foreground">Colors</span>
              <div className="flex items-center gap-2">
                <div className="h-6 w-6 rounded-full border border-border" style={{ backgroundColor: formData.primary_color }} />
                <div className="h-6 w-6 rounded-full border border-border" style={{ backgroundColor: formData.secondary_color }} />
              </div>
            </div>
          </div>
        </div>

        {/* Contact */}
        <div className="bg-muted/30 rounded-lg border border-border p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Globe className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold text-foreground">Contact Info</h3>
            </div>
            <button onClick={() => onEditStep(2)} className="text-xs font-medium text-primary hover:text-primary/80" aria-label="Edit contact info">
              Edit
            </button>
          </div>
          <div className="divide-y divide-border">
            <div className="flex items-center gap-2 py-2">
              <Mail className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-sm text-foreground">{formData.contact_email || '-'}</span>
            </div>
            <div className="flex items-center gap-2 py-2">
              <Phone className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-sm text-foreground">{formData.contact_phone || '-'}</span>
            </div>
            <div className="flex items-center gap-2 py-2">
              <MapPin className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-sm text-foreground">{formData.address || '-'}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="flex justify-between pt-4">
        <button
          onClick={onBack}
          disabled={isSubmitting}
          className="inline-flex items-center rounded-md border border-input bg-background px-6 py-2.5 text-sm font-medium text-foreground hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50 transition-colors"
        >
          Back
        </button>
        <button
          onClick={onSubmit}
          disabled={isSubmitting}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50 transition-colors"
        >
          {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
          {isSubmitting ? 'Creating...' : 'Create Dealer'}
        </button>
      </div>
    </div>
  )
}
