import { useState } from 'react'
import type { WizardFormData } from '../types'
import DeviceFrame from './DeviceFrame'
import LoginPagePreview from './LoginPagePreview'
import MobileHomePreview from './MobileHomePreview'

interface Props {
  formData: WizardFormData
}

export default function DealerPreviewPanel({ formData }: Props) {
  const [tab, setTab] = useState<'login' | 'home'>('login')

  return (
    <div className="flex flex-col items-center">
      <div className="flex items-center gap-1 mb-6 bg-muted rounded-lg p-1">
        <button
          onClick={() => setTab('login')}
          className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
            tab === 'login' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Login
        </button>
        <button
          onClick={() => setTab('home')}
          className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
            tab === 'home' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          App Home
        </button>
      </div>

      <DeviceFrame>
        {tab === 'login' ? (
          <LoginPagePreview formData={formData} />
        ) : (
          <MobileHomePreview formData={formData} />
        )}
      </DeviceFrame>
    </div>
  )
}
