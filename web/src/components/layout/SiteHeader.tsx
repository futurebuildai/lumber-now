import { useAuth } from '@/context/AuthContext'
import { useNavigate } from 'react-router-dom'
import { LogOut, User } from 'lucide-react'

export default function SiteHeader({ title }: { title?: string }) {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <header className="h-16 border-b border-border bg-card flex items-center justify-between px-6" role="banner" aria-label="Site header">
      <h1 className="text-lg font-semibold text-foreground">{title || 'Dashboard'}</h1>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <User className="h-4 w-4" />
          <span>{user?.full_name}</span>
          <span className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded-full">
            {user?.role.replace('_', ' ')}
          </span>
        </div>
        <button
          onClick={() => { logout(); navigate('/login') }}
          className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1"
          aria-label="Sign out"
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </button>
      </div>
    </header>
  )
}
