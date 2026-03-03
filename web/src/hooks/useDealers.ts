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

export function useGetDealer(id: string) {
  return useQuery({
    queryKey: ['dealers', id],
    queryFn: async () => {
      const { data } = await api.get<Dealer>(`/platform/dealers/${id}`)
      return data
    },
    enabled: !!id,
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

export function useUpdateDealer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...dealer }: Partial<Dealer> & { id: string }) => {
      const { data } = await api.put<Dealer>(`/platform/dealers/${id}`, dealer)
      return data
    },
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['dealers'] })
      qc.invalidateQueries({ queryKey: ['dealers', variables.id] })
    },
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
