import { useQuery } from '@tanstack/react-query'
import api from '@/api/client'
import type { User } from '@/types'

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const { data } = await api.get<{ users: User[] }>('/admin/users')
      return data.users || []
    },
  })
}
