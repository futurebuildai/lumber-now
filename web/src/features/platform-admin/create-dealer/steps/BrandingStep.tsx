import { useRef } from 'react'
import type { WizardFormData } from '../types'
import { useUploadLogo } from '../hooks/useUploadLogo'
import { Upload, X, Palette, Loader2 } from 'lucide-react'

interface Props {
  formData: WizardFormData
  onUpdate: (data: Partial<WizardFormData>) => void
  onNext: () => void
  onBack: () => void
}

export default function BrandingStep({ formData, onUpdate, onNext, onBack }: Props) {
  const fileRef = useRef<HTMLInputElement>(null)
  const uploadLogo = useUploadLogo()

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    try {
      const result = await uploadLogo.mutateAsync(file)
      onUpdate({ logo_url: result.url, logo_key: result.key })
    } catch {
      // Error handled by mutation state
    }
  }

  const removeLogo = () => {
    onUpdate({ logo_url: '', logo_key: '' })
    if (fileRef.current) fileRef.current.value = ''
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-foreground">Branding</h2>
        <p className="text-sm text-muted-foreground mt-1">Upload a logo and choose brand colors. See the preview on the right.</p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground">Logo</label>
          {formData.logo_url ? (
            <div className="flex items-center gap-4 p-4 bg-muted/50 rounded-lg border border-border">
              <img src={formData.logo_url} alt="Logo preview" className="h-16 w-16 object-contain rounded-md" />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-foreground truncate">Logo uploaded</p>
                <p className="text-xs text-muted-foreground truncate">{formData.logo_key}</p>
              </div>
              <button
                onClick={removeLogo}
                className="p-1.5 rounded-md hover:bg-background text-muted-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          ) : (
            <div
              onClick={() => fileRef.current?.click()}
              className="flex flex-col items-center justify-center p-8 border-2 border-dashed border-border rounded-lg cursor-pointer hover:border-primary/50 hover:bg-muted/50 transition-colors"
            >
              {uploadLogo.isPending ? (
                <Loader2 className="h-8 w-8 text-muted-foreground animate-spin" />
              ) : (
                <Upload className="h-8 w-8 text-muted-foreground" />
              )}
              <p className="mt-2 text-sm text-muted-foreground">
                {uploadLogo.isPending ? 'Uploading...' : 'Click to upload logo'}
              </p>
              <p className="text-xs text-muted-foreground mt-1">PNG, JPG, or SVG</p>
            </div>
          )}
          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            onChange={handleFileSelect}
            className="hidden"
          />
        </div>

        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Palette className="h-4 w-4 text-muted-foreground" />
            <label className="text-sm font-medium text-foreground">Brand Colors</label>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-xs text-muted-foreground">Primary Color</label>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={formData.primary_color}
                  onChange={(e) => onUpdate({ primary_color: e.target.value })}
                  className="h-10 w-14 rounded-md border border-input cursor-pointer"
                />
                <input
                  type="text"
                  value={formData.primary_color}
                  onChange={(e) => onUpdate({ primary_color: e.target.value })}
                  className="flex h-10 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-mono"
                />
              </div>
            </div>
            <div className="space-y-1">
              <label className="text-xs text-muted-foreground">Secondary Color</label>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={formData.secondary_color}
                  onChange={(e) => onUpdate({ secondary_color: e.target.value })}
                  className="h-10 w-14 rounded-md border border-input cursor-pointer"
                />
                <input
                  type="text"
                  value={formData.secondary_color}
                  onChange={(e) => onUpdate({ secondary_color: e.target.value })}
                  className="flex h-10 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-mono"
                />
              </div>
            </div>
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
