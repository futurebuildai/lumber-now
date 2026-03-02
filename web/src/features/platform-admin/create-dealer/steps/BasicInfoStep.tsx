import type { WizardFormData } from '../types'

interface Props {
  formData: WizardFormData
  onUpdate: (data: Partial<WizardFormData>) => void
  onNext: () => void
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export default function BasicInfoStep({ formData, onUpdate, onNext }: Props) {
  const handleNameChange = (name: string) => {
    const slug = slugify(name)
    onUpdate({
      name,
      slug,
      subdomain: slug ? `${slug}.lumbernow.com` : '',
    })
  }

  const isValid = formData.name.trim() && formData.slug.trim()

  return (
    <div role="form" aria-label="Basic information" className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-foreground">Basic Information</h2>
        <p id="basic-info-desc" className="text-sm text-muted-foreground mt-1">Enter the dealer name. The slug and subdomain will be generated automatically.</p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <label htmlFor="dealer-name" className="text-sm font-medium text-foreground">Dealer Name *</label>
          <input
            id="dealer-name"
            type="text"
            value={formData.name}
            onChange={(e) => handleNameChange(e.target.value)}
            placeholder="Acme Lumber Co."
            aria-required="true"
            aria-describedby="basic-info-desc"
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <label htmlFor="dealer-slug" className="text-sm font-medium text-foreground">Slug *</label>
            <input
              id="dealer-slug"
              type="text"
              value={formData.slug}
              onChange={(e) => onUpdate({ slug: e.target.value, subdomain: `${e.target.value}.lumbernow.com` })}
              placeholder="acme-lumber"
              aria-required="true"
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
          </div>
          <div className="space-y-2">
            <label htmlFor="dealer-subdomain" className="text-sm font-medium text-foreground">Subdomain</label>
            <input
              id="dealer-subdomain"
              type="text"
              value={formData.subdomain}
              readOnly
              aria-readonly="true"
              className="flex h-10 w-full rounded-md border border-input bg-muted px-3 py-2 text-sm text-muted-foreground"
            />
          </div>
        </div>
      </div>

      <div className="flex justify-end pt-4">
        <button
          onClick={onNext}
          disabled={!isValid}
          className="inline-flex items-center rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50 transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  )
}
