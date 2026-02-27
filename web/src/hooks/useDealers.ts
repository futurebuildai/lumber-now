import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/api/client'
import type { Dealer } from '@/types'

export function useDealers() {
  return useQuery({
    queryKey: ['dealers'],
    queryFn: async () => {
      const { data } = await api.get<{ dealers: Dealer[] }>('/platform/dealers')
      return data.dealers || []
    },
  })
}

export function useCreateDealer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (dealer: Partial<Dealer>) => {
      const { data } = await api.post<Dealer>('/platform/dealers', dealer)
      return data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['dealers'] }),
  })
}

export function useToggleDealerActive() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, active }: { id: string; active: boolean }) => {
      await api.post(`/platform/dealers/${id}/${active ? 'deactivate' : 'activate'}`)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['dealers'] }),
  })
}
