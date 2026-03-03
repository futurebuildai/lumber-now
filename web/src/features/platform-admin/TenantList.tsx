import { Link } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { useDealers, useToggleDealerActive } from '@/hooks/useDealers'
import api from '@/api/client'
import { Building2, Plus, Pencil, Smartphone } from 'lucide-react'

function useTriggerBuild() {
  return useMutation({
    mutationFn: async ({ dealerSlug, platform }: { dealerSlug: string; platform: string }) => {
      const { data } = await api.post('/platform/builds', { dealer_slug: dealerSlug, platform })
      return data
    },
  })
}

export default function TenantList() {
  const { data: dealers, isLoading } = useDealers()
  const toggleActive = useToggleDealerActive()
  const triggerBuild = useTriggerBuild()

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div className="h-8 w-48 bg-muted animate-pulse rounded" />
          <div className="h-9 w-32 bg-muted animate-pulse rounded" />
        </div>
        <div className="bg-card border border-border rounded-lg overflow-hidden">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="flex items-center gap-4 px-6 py-4 border-b border-border last:border-0">
              <div className="h-4 w-32 bg-muted animate-pulse rounded" />
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
              <div className="h-4 w-40 bg-muted animate-pulse rounded" />
              <div className="h-5 w-16 bg-muted animate-pulse rounded-full" />
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-foreground">Platform: Dealers</h1>
        <Link
          to="/platform/new-dealer"
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 transition-colors"
        >
          <Plus className="h-4 w-4" />
          New Dealer
        </Link>
      </div>

      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted/50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Slug</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Subdomain</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {dealers?.map((d) => (
              <tr key={d.id} className="hover:bg-muted/50 transition-colors">
                <td className="px-6 py-4 text-sm font-medium text-foreground">{d.name}</td>
                <td className="px-6 py-4 text-sm text-muted-foreground font-mono">{d.slug}</td>
                <td className="px-6 py-4 text-sm text-muted-foreground">{d.subdomain}</td>
                <td className="px-6 py-4">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${d.active ? 'bg-green-500/10 text-green-700 dark:text-green-400' : 'bg-destructive/10 text-destructive'}`}>
                    {d.active ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm">
                  <div className="flex items-center gap-3">
                    <Link
                      to={`/platform/dealers/${d.id}/edit`}
                      className="inline-flex items-center gap-1 font-medium text-primary hover:text-primary/80"
                    >
                      <Pencil className="h-3.5 w-3.5" />
                      Edit
                    </Link>
                    <button
                      onClick={() => triggerBuild.mutate({ dealerSlug: d.slug, platform: 'both' })}
                      disabled={triggerBuild.isPending}
                      className="inline-flex items-center gap-1 font-medium text-muted-foreground hover:text-foreground disabled:opacity-50"
                    >
                      <Smartphone className="h-3.5 w-3.5" />
                      Build App
                    </button>
                    <button
                      onClick={() => toggleActive.mutate({ id: d.id, active: d.active })}
                      disabled={toggleActive.isPending}
                      className={`font-medium disabled:opacity-50 ${d.active ? 'text-destructive hover:text-destructive/80' : 'text-green-700 hover:text-green-600'}`}
                    >
                      {d.active ? 'Deactivate' : 'Activate'}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {(!dealers || dealers.length === 0) && (
              <tr>
                <td colSpan={5} className="px-6 py-16 text-center">
                  <Building2 className="h-10 w-10 text-muted-foreground/50 mx-auto mb-3" />
                  <p className="text-sm font-medium text-foreground">No dealers</p>
                  <p className="text-sm text-muted-foreground mt-1">Create your first dealer to get started.</p>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
