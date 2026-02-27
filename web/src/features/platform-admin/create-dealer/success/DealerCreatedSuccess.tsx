import { Link } from 'react-router-dom'
import type { WizardFormData } from '../types'
import CreateAdminUserForm from './CreateAdminUserForm'
import { CheckCircle, Globe, ArrowLeft } from 'lucide-react'

interface Props {
  formData: WizardFormData
  dealerId: string
}

export default function DealerCreatedSuccess({ formData, dealerId }: Props) {
  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex flex-col items-center text-center py-6">
        <div className="h-16 w-16 rounded-full bg-green-500/10 flex items-center justify-center mb-4">
          <CheckCircle className="h-8 w-8 text-green-600" />
        </div>
        <h1 className="text-2xl font-bold text-foreground">Dealer Created!</h1>
        <p className="text-sm text-muted-foreground mt-2">
          <strong>{formData.name}</strong> has been set up successfully.
        </p>
      </div>

      {/* URL info */}
      <div className="bg-card border border-border rounded-lg p-4">
        <div className="flex items-center gap-2 mb-2">
          <Globe className="h-4 w-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold text-foreground">Web Portal URL</h3>
        </div>
        <div className="bg-muted rounded-md px-3 py-2">
          <code className="text-sm text-foreground">{formData.subdomain}</code>
        </div>
      </div>

      {/* Create admin user */}
      <div className="bg-card border border-border rounded-lg p-4">
        <CreateAdminUserForm dealerId={dealerId} />
      </div>

      <div className="flex justify-center pt-4">
        <Link
          to="/platform"
          className="inline-flex items-center gap-2 rounded-md border border-input bg-background px-6 py-2.5 text-sm font-medium text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Dealers
        </Link>
      </div>
    </div>
  )
}
