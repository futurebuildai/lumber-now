import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'
import { useTenant } from '@/hooks/useTenant'
import { cn } from '@/lib/utils'
import {
  ClipboardList,
  Package,
  Users,
  Settings,
  Building2,
  LogOut,
  ChevronLeft,
  Menu,
} from 'lucide-react'
import { useState } from 'react'

const navItems = [
  { to: '/requests', icon: ClipboardList, label: 'Requests', roles: null },
  { to: '/inventory', icon: Package, label: 'Inventory', roles: ['dealer_admin', 'sales_rep', 'platform_admin'] },
  { to: '/admin/users', icon: Users, label: 'Users', roles: ['dealer_admin', 'platform_admin'] },
  { to: '/admin/settings', icon: Settings, label: 'Settings', roles: ['dealer_admin', 'platform_admin'] },
  { to: '/platform', icon: Building2, label: 'Platform', roles: ['platform_admin'] },
]

export default function AppSidebar() {
  const { user, logout } = useAuth()
  const { config } = useTenant()
  const location = useLocation()
  const navigate = useNavigate()
  const [collapsed, setCollapsed] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const initials = user?.full_name
    ?.split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2) || '?'

  return (
    <aside
      className={cn(
        'flex flex-col h-screen bg-sidebar-background border-r border-sidebar-border transition-all duration-200',
        collapsed ? 'w-16' : 'w-64'
      )}
    >
      {/* Header */}
      <div className="flex items-center h-16 px-4 border-b border-sidebar-border">
        {!collapsed && (
          <div className="flex items-center gap-2 flex-1 min-w-0">
            {config?.logo_url ? (
              <img src={config.logo_url} alt="" className="h-8 w-8 object-contain flex-shrink-0" />
            ) : (
              <Building2 className="h-6 w-6 text-sidebar-primary flex-shrink-0" />
            )}
            <span className="font-semibold text-sidebar-foreground truncate">
              {config?.name || 'LumberNow'}
            </span>
          </div>
        )}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="p-1.5 rounded-md hover:bg-sidebar-accent text-sidebar-foreground"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <Menu className="h-5 w-5" /> : <ChevronLeft className="h-5 w-5" />}
        </button>
      </div>

      {/* Nav */}
      <nav className="flex-1 py-4 px-2 space-y-1 overflow-y-auto">
        {navItems
          .filter(item => !item.roles || (user && item.roles.includes(user.role)))
          .map(item => {
            const active = location.pathname.startsWith(item.to)
            return (
              <Link
                key={item.to}
                to={item.to}
                className={cn(
                  'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                  active
                    ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                    : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
                )}
                title={collapsed ? item.label : undefined}
              >
                <item.icon className="h-5 w-5 flex-shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </Link>
            )
          })}
      </nav>

      {/* User footer */}
      <div className="border-t border-sidebar-border p-3">
        <div className={cn('flex items-center', collapsed ? 'justify-center' : 'gap-3')}>
          <div className="h-8 w-8 rounded-full bg-sidebar-primary text-sidebar-primary-foreground flex items-center justify-center text-xs font-semibold flex-shrink-0">
            {initials}
          </div>
          {!collapsed && (
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-sidebar-foreground truncate">{user?.full_name}</p>
              <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
            </div>
          )}
          {!collapsed && (
            <button
              onClick={handleLogout}
              className="p-1.5 rounded-md hover:bg-sidebar-accent text-muted-foreground"
              aria-label="Sign out"
            >
              <LogOut className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>
    </aside>
  )
}
