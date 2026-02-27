import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useRequests, useCreateRequest, useProcessRequest } from '@/hooks/useRequests'
import StatusBadge from '@/components/ui/StatusBadge'
import { Plus, X, ClipboardList, Send } from 'lucide-react'

export default function RequestFeed() {
  const [showCreate, setShowCreate] = useState(false)
  const [rawText, setRawText] = useState('')
  const { data: requests, isLoading } = useRequests()
  const createRequest = useCreateRequest()
  const processRequest = useProcessRequest()

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    await createRequest.mutateAsync({ input_type: 'text', raw_text: rawText })
    setRawText('')
    setShowCreate(false)
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <div className="h-8 w-48 bg-muted animate-pulse rounded" />
          <div className="h-9 w-32 bg-muted animate-pulse rounded" />
        </div>
        <div className="bg-card border border-border rounded-lg overflow-hidden">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="flex items-center gap-4 px-6 py-4 border-b border-border last:border-0">
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
              <div className="h-4 w-16 bg-muted animate-pulse rounded" />
              <div className="h-5 w-20 bg-muted animate-pulse rounded-full" />
              <div className="h-4 w-12 bg-muted animate-pulse rounded" />
              <div className="h-4 w-24 bg-muted animate-pulse rounded" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-foreground">Material Requests</h1>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 transition-colors"
        >
          {showCreate ? <X className="h-4 w-4" /> : <Plus className="h-4 w-4" />}
          {showCreate ? 'Cancel' : 'New Request'}
        </button>
      </div>

      {showCreate && (
        <form onSubmit={handleCreate} className="bg-card border border-border rounded-lg p-6 space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground">Describe your materials needed</label>
            <textarea
              value={rawText}
              onChange={(e) => setRawText(e.target.value)}
              rows={4}
              className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              placeholder="e.g., I need 100 2x4x8 studs, 50 sheets of 1/2 inch OSB, and 20 tubes of PL Premium..."
            />
          </div>
          <div className="flex gap-3">
            <button
              type="submit"
              disabled={createRequest.isPending || !rawText.trim()}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50 transition-colors"
            >
              <Send className="h-4 w-4" />
              {createRequest.isPending ? 'Submitting...' : 'Submit Request'}
            </button>
            <button
              type="button"
              onClick={() => setShowCreate(false)}
              className="inline-flex items-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted/50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">ID</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Type</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Confidence</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Created</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {requests?.map((req) => (
              <tr key={req.id} className="hover:bg-muted/50 transition-colors">
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <Link to={`/requests/${req.id}`} className="font-medium text-primary hover:text-primary/80">
                    {req.id.slice(0, 8)}...
                  </Link>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground capitalize">{req.input_type}</td>
                <td className="px-6 py-4 whitespace-nowrap"><StatusBadge status={req.status} /></td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                  {req.ai_confidence ? `${(parseFloat(req.ai_confidence) * 100).toFixed(0)}%` : '-'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                  {new Date(req.created_at).toLocaleDateString()}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  {req.status === 'pending' && (
                    <button
                      onClick={() => processRequest.mutate(req.id)}
                      disabled={processRequest.isPending}
                      className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:text-primary/80 disabled:opacity-50"
                    >
                      Process
                    </button>
                  )}
                </td>
              </tr>
            ))}
            {(!requests || requests.length === 0) && (
              <tr>
                <td colSpan={6} className="px-6 py-16 text-center">
                  <ClipboardList className="h-10 w-10 text-muted-foreground/50 mx-auto mb-3" />
                  <p className="text-sm font-medium text-foreground">No requests yet</p>
                  <p className="text-sm text-muted-foreground mt-1">Create your first material request to get started.</p>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
