import type { WizardFormData } from '../types'
import { Building2, Mail, Lock } from 'lucide-react'

interface Props {
  formData: WizardFormData
}

export default function LoginPagePreview({ formData }: Props) {
  const { name, logo_url, primary_color } = formData

  return (
    <div className="flex flex-col items-center justify-center h-full px-6 bg-background">
      {/* Logo */}
      <div className="mb-4">
        {logo_url ? (
          <img src={logo_url} alt="" className="h-12 w-12 object-contain" />
        ) : (
          <div
            className="h-12 w-12 rounded-xl flex items-center justify-center"
            style={{ backgroundColor: primary_color }}
          >
            <Building2 className="h-6 w-6 text-white" />
          </div>
        )}
      </div>

      {/* Title */}
      <h2 className="text-sm font-bold text-foreground mb-0.5">{name || 'Dealer Name'}</h2>
      <p className="text-[10px] text-muted-foreground mb-4">Sign in to your account</p>

      {/* Mock form */}
      <div className="w-full space-y-2">
        <div className="flex items-center gap-1.5 h-7 rounded border border-border px-2 bg-muted/30">
          <Mail className="h-3 w-3 text-muted-foreground" />
          <span className="text-[10px] text-muted-foreground">you@example.com</span>
        </div>
        <div className="flex items-center gap-1.5 h-7 rounded border border-border px-2 bg-muted/30">
          <Lock className="h-3 w-3 text-muted-foreground" />
          <span className="text-[10px] text-muted-foreground">Password</span>
        </div>
        <div
          className="h-7 rounded flex items-center justify-center"
          style={{ backgroundColor: primary_color }}
        >
          <span className="text-[10px] font-medium text-white">Sign in</span>
        </div>
      </div>

      <p className="text-[9px] text-muted-foreground mt-3">
        Don't have an account? <span style={{ color: primary_color }}>Register</span>
      </p>
    </div>
  )
}
