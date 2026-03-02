import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import api from '@/api/client'
import { User, Mail, Lock, Phone, CheckCircle, AlertCircle, Loader2 } from 'lucide-react'

interface Props {
  dealerId: string
}

export default function CreateAdminUserForm({ dealerId }: Props) {
  const [form, setForm] = useState({
    email: '',
    password: '',
    full_name: '',
    phone: '',
  })
  const [created, setCreated] = useState(false)

  const createUser = useMutation({
    mutationFn: async () => {
      const { data } = await api.post(`/platform/dealers/${dealerId}/users`, form)
      return data
    },
    onSuccess: () => setCreated(true),
  })

  if (created) {
    return (
      <div className="flex items-center gap-2 bg-green-500/10 text-green-700 dark:text-green-400 px-4 py-3 rounded-md text-sm">
        <CheckCircle className="h-4 w-4 flex-shrink-0" />
        Admin user created for {form.email}
      </div>
    )
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createUser.mutate()
  }

  const isValid = form.email && form.password.length >= 8 && form.full_name

  return (
    <form onSubmit={handleSubmit} role="form" aria-label="Create first admin user" className="space-y-4">
      <h3 className="text-sm font-semibold text-foreground">Create First Admin User</h3>

      {createUser.isError && (
        <div role="alert" aria-live="assertive" className="flex items-center gap-2 bg-destructive/10 text-destructive px-4 py-3 rounded-md text-sm">
          <AlertCircle className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
          Failed to create user. Email may already be in use.
        </div>
      )}

      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1">
          <label htmlFor="admin-full-name" className="text-xs font-medium text-foreground">Full Name *</label>
          <div className="relative">
            <User className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" aria-hidden="true" />
            <input
              id="admin-full-name"
              type="text"
              required
              aria-required="true"
              value={form.full_name}
              onChange={(e) => setForm({ ...form, full_name: e.target.value })}
              placeholder="Jane Smith"
              className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 pl-8 text-sm"
            />
          </div>
        </div>
        <div className="space-y-1">
          <label htmlFor="admin-email" className="text-xs font-medium text-foreground">Email *</label>
          <div className="relative">
            <Mail className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" aria-hidden="true" />
            <input
              id="admin-email"
              type="email"
              required
              aria-required="true"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder="admin@acmelumber.com"
              className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 pl-8 text-sm"
            />
          </div>
        </div>
        <div className="space-y-1">
          <label htmlFor="admin-password" className="text-xs font-medium text-foreground">Password * (min 8 chars)</label>
          <div className="relative">
            <Lock className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" aria-hidden="true" />
            <input
              id="admin-password"
              type="password"
              required
              aria-required="true"
              aria-describedby="password-hint"
              minLength={8}
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 pl-8 text-sm"
            />
            <span id="password-hint" className="sr-only">Minimum 8 characters</span>
          </div>
        </div>
        <div className="space-y-1">
          <label htmlFor="admin-phone" className="text-xs font-medium text-foreground">Phone</label>
          <div className="relative">
            <Phone className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" aria-hidden="true" />
            <input
              id="admin-phone"
              type="tel"
              value={form.phone}
              onChange={(e) => setForm({ ...form, phone: e.target.value })}
              className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 pl-8 text-sm"
            />
          </div>
        </div>
      </div>

      <button
        type="submit"
        disabled={!isValid || createUser.isPending}
        className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50 transition-colors"
      >
        {createUser.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
        {createUser.isPending ? 'Creating...' : 'Create Admin User'}
      </button>
    </form>
  )
}
