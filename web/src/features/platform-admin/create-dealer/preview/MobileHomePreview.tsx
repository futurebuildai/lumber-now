import type { WizardFormData } from '../types'
import { Building2, Mic, Camera, FileText, MessageSquare, Clock } from 'lucide-react'

interface Props {
  formData: WizardFormData
}

export default function MobileHomePreview({ formData }: Props) {
  const { name, logo_url, primary_color, secondary_color } = formData

  return (
    <div className="flex flex-col h-full bg-background">
      {/* App bar */}
      <div
        className="flex items-center gap-2 px-3 py-2"
        style={{ backgroundColor: primary_color }}
      >
        {logo_url ? (
          <img src={logo_url} alt="" className="h-5 w-5 object-contain rounded" />
        ) : (
          <Building2 className="h-4 w-4 text-white" />
        )}
        <span className="text-[11px] font-semibold text-white truncate">{name || 'Dealer Name'}</span>
      </div>

      <div className="flex-1 px-3 py-3 space-y-2.5 overflow-hidden">
        {/* Welcome card */}
        <div className="rounded-lg p-2.5" style={{ backgroundColor: `${primary_color}10` }}>
          <p className="text-[10px] text-muted-foreground">Welcome back</p>
          <p className="text-[11px] font-semibold text-foreground">John Doe</p>
        </div>

        {/* New Request hero */}
        <div
          className="rounded-lg p-3 text-white"
          style={{ background: `linear-gradient(135deg, ${primary_color}, ${secondary_color})` }}
        >
          <p className="text-[11px] font-bold mb-1">New Request</p>
          <p className="text-[9px] opacity-80">Submit a material order</p>
          <div className="flex gap-1.5 mt-2">
            {[
              { icon: MessageSquare, label: 'Text' },
              { icon: Mic, label: 'Voice' },
              { icon: Camera, label: 'Photo' },
              { icon: FileText, label: 'PDF' },
            ].map(({ icon: Icon, label }) => (
              <div key={label} className="flex flex-col items-center gap-0.5 bg-white/20 rounded px-1.5 py-1">
                <Icon className="h-3 w-3" />
                <span className="text-[8px]">{label}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Recent requests */}
        <div>
          <p className="text-[10px] font-semibold text-foreground mb-1.5">Recent Requests</p>
          {[
            { status: 'Confirmed', time: '2h ago', items: 5 },
            { status: 'Processing', time: '1d ago', items: 3 },
          ].map((req, i) => (
            <div key={i} className="flex items-center justify-between py-1.5 border-b border-border last:border-0">
              <div className="flex items-center gap-1.5">
                <Clock className="h-3 w-3 text-muted-foreground" />
                <div>
                  <p className="text-[9px] font-medium text-foreground">{req.items} items</p>
                  <p className="text-[8px] text-muted-foreground">{req.time}</p>
                </div>
              </div>
              <span
                className="text-[8px] px-1.5 py-0.5 rounded-full font-medium"
                style={{
                  backgroundColor: `${primary_color}15`,
                  color: primary_color,
                }}
              >
                {req.status}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
