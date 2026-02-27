import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/api/client'

interface DealerSettingsData {
  name: string
  logo_url: string
  primary_color: string
  secondary_color: string
  contact_email: string
  contact_phone: string
  address: string
}

export function useDealerSettings() {
  return useQuery({
    queryKey: ['dealer-settings'],
    queryFn: async () => {
      const { data } = await api.get<DealerSettingsData>('/admin/settings')
      return data
    },
  })
}

export function useUpdateDealerSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (settings: DealerSettingsData) => {
      const { data } = await api.put('/admin/settings', settings)
      return data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['dealer-settings'] }),
  })
}
