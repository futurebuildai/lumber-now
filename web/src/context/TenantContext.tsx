import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import api from '@/api/client'
import type { TenantConfig } from '@/types'
import { hexToHsl, contrastForeground } from '@/lib/color-utils'

interface TenantState {
  config: TenantConfig | null
  loading: boolean
  error: string | null
}

const TenantContext = createContext<TenantState>({
  config: null,
  loading: true,
  error: null,
})

function getSlug(): string {
  const hostname = window.location.hostname
  // localhost -> use env var or "demo"
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return import.meta.env.VITE_DEFAULT_TENANT_SLUG || 'demo'
  }
  // acme.lumbernow.com -> acme
  const parts = hostname.split('.')
  if (parts.length >= 3) return parts[0]
  return import.meta.env.VITE_DEFAULT_TENANT_SLUG || 'demo'
}

function injectCssVars(config: TenantConfig) {
  const root = document.documentElement.style
  const primaryHsl = hexToHsl(config.primary_color)
  const secondaryHsl = hexToHsl(config.secondary_color)

  if (primaryHsl) {
    root.setProperty('--primary', `${primaryHsl.h} ${primaryHsl.s}% ${primaryHsl.l}%`)
    root.setProperty('--primary-foreground', contrastForeground(config.primary_color) === '#ffffff' ? '0 0% 100%' : '0 0% 0%')
    root.setProperty('--ring', `${primaryHsl.h} ${primaryHsl.s}% ${primaryHsl.l}%`)
    root.setProperty('--sidebar-primary', `${primaryHsl.h} ${primaryHsl.s}% ${primaryHsl.l}%`)
    root.setProperty('--sidebar-primary-foreground', contrastForeground(config.primary_color) === '#ffffff' ? '0 0% 100%' : '0 0% 0%')
  }
  if (secondaryHsl) {
    root.setProperty('--secondary', `${secondaryHsl.h} ${secondaryHsl.s}% ${secondaryHsl.l}%`)
    root.setProperty('--secondary-foreground', contrastForeground(config.secondary_color) === '#ffffff' ? '0 0% 100%' : '0 0% 0%')
  }
}

export function TenantProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<TenantState>({
    config: null,
    loading: true,
    error: null,
  })

  useEffect(() => {
    const slug = getSlug()
    if (!slug) {
      setState({ config: null, loading: false, error: 'No tenant configured' })
      return
    }

    api.get<TenantConfig>('/tenant/config', { params: { slug } })
      .then(({ data }) => {
        localStorage.setItem('tenant_id', data.dealer_id)
        injectCssVars(data)
        setState({ config: data, loading: false, error: null })
      })
      .catch(() => {
        setState({ config: null, loading: false, error: `Tenant "${slug}" not found` })
      })
  }, [])

  return (
    <TenantContext.Provider value={state}>
      {children}
    </TenantContext.Provider>
  )
}

export function useTenant() {
  return useContext(TenantContext)
}
