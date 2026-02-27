import type { RequestStatus } from '@/types'

const statusStyles: Record<RequestStatus, string> = {
  pending: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400',
  processing: 'bg-blue-500/10 text-blue-700 dark:text-blue-400',
  parsed: 'bg-purple-500/10 text-purple-700 dark:text-purple-400',
  confirmed: 'bg-green-500/10 text-green-700 dark:text-green-400',
  sent: 'bg-muted text-muted-foreground',
  failed: 'bg-destructive/10 text-destructive',
}

export default function StatusBadge({ status }: { status: RequestStatus }) {
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${statusStyles[status]}`}>
      {status}
    </span>
  )
}
